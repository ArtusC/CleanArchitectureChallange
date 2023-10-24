package usecase

import (
	"github.com/ArtusC/CleanArchitectureChallange/internal/entity"
)

type ListOrdersOutputDTO struct {
	ID         string  `json:"id"`
	Price      float64 `json:"price"`
	Tax        float64 `json:"tax"`
	FinalPrice float64 `json:"final_price"`
}

type ListOrdersUseCase struct {
	OrderRepository entity.OrderRepositoryInterface
}

func NewListOrdersUseCase(
	OrderRepository entity.OrderRepositoryInterface,
) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		OrderRepository: OrderRepository,
	}
}

func (l *ListOrdersUseCase) Execute() ([]ListOrdersOutputDTO, error) {
	orders, err := l.OrderRepository.GetOrders()
	if err != nil {
		return nil, err
	}

	var ordersResponse []ListOrdersOutputDTO

	for _, order := range orders {
		orderResponse := ListOrdersOutputDTO{
			ID:         order.ID,
			Price:      order.Price,
			Tax:        order.Tax,
			FinalPrice: order.FinalPrice,
		}

		ordersResponse = append(ordersResponse, orderResponse)
	}

	return ordersResponse, nil
}
