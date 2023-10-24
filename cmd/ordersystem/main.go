package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/ArtusC/CleanArchitectureChallange/configs"
	"github.com/ArtusC/CleanArchitectureChallange/internal/event/handler"
	"github.com/ArtusC/CleanArchitectureChallange/internal/infra/graph"
	"github.com/ArtusC/CleanArchitectureChallange/internal/infra/grpc/pb"
	"github.com/ArtusC/CleanArchitectureChallange/internal/infra/grpc/service"
	"github.com/ArtusC/CleanArchitectureChallange/internal/infra/web/webserver"
	"github.com/ArtusC/CleanArchitectureChallange/pkg/events"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	// mysql
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(fmt.Sprintf("LoadConfigError: %s", err.Error()))
	}

	db, err := sql.Open(configs.DBDriver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", configs.DBUser, configs.DBPassword, configs.DBHost, configs.DBPort, configs.DBName))
	if err != nil {
		panic(fmt.Sprintf("sqlOpenError: %s", err.Error()))
	}
	defer db.Close()

	rabbitMQChannel := getRabbitMQChannel()

	eventDispatcher := events.NewEventDispatcher()
	if err := eventDispatcher.Register("OrderCreated", &handler.OrderCreatedHandler{
		RabbitMQChannel: rabbitMQChannel,
	}); err != nil || eventDispatcher == nil {
		panic(fmt.Sprintf("eventDispatcherNIL: %s", err.Error()))
	}

	if db == nil {
		panic(fmt.Sprintf("dbNIL: %s", err.Error()))
	}

	createOrderUseCase := NewCreateOrderUseCase(db, eventDispatcher)
	listOrderUseCase := NewListOrdersUseCase(db)

	webserver := webserver.NewWebServer(configs.WebServerPort)
	webOrderHandler := NewWebOrderHandler(db, eventDispatcher)
	webserver.AddHandler("/order", webOrderHandler.Create)
	webserver.AddHandler("/list-orders", webOrderHandler.List)
	fmt.Println("Starting web server on port", configs.WebServerPort)
	go webserver.Start()

	grpcServer := grpc.NewServer()
	createOrderService := service.NewOrderService(*createOrderUseCase, *listOrderUseCase)
	pb.RegisterOrderServiceServer(grpcServer, createOrderService)
	reflection.Register(grpcServer)

	fmt.Println("Starting gRPC server on port", configs.GRPCServerPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", configs.GRPCServerPort))
	if err != nil {
		panic(fmt.Sprintf("grpcListenError: %s", err.Error()))
	}
	go grpcServer.Serve(lis)

	srv := graphql_handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		CreateOrderUseCase: *createOrderUseCase,
		ListOrdersUseCase:  *listOrderUseCase,
	}}))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	fmt.Println("Starting GraphQL server on port", configs.GraphQLServerPort)
	http.ListenAndServe(":"+configs.GraphQLServerPort, nil)
}

func getRabbitMQChannel() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(fmt.Sprintf("connRabbitError: %s", err.Error()))
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("channelRAbbitError: %s", err.Error()))
	}

	q, err := ch.QueueDeclare(
		"orders", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		panic(fmt.Sprintf("queueDeclarationRabbitError: %s", err.Error()))
	}

	err = ch.QueueBind(
		q.Name,       // queue name
		"",           // routing key
		"amq.direct", // exchange
		false,
		nil,
	)
	if err != nil {
		panic(fmt.Sprintf("queueBindRabbitError: %s", err.Error()))
	}
	return ch
}
