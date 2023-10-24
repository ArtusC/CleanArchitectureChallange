package entity

type OrderRepositoryInterface interface {
	Save(order *Order) error
	GetOrders() ([]Order, error)
}
