package utils

import (
	"github.com/wassef911/astore/internal/api/dto"
	"github.com/wassef911/astore/internal/delivery/aggregate"
	"github.com/wassef911/astore/internal/delivery/models"
)

func OrderProjectionFrom(orderAggregate *aggregate.OrderAggregate) *models.OrderProjection {
	return &models.OrderProjection{
		OrderID:         aggregate.GetOrderAggregateID(orderAggregate.GetID()),
		ShopItems:       orderAggregate.Order.ShopItems,
		Paid:            orderAggregate.Order.Paid,
		Submitted:       orderAggregate.Order.Submitted,
		Completed:       orderAggregate.Order.Completed,
		Canceled:        orderAggregate.Order.Canceled,
		AccountEmail:    orderAggregate.Order.AccountEmail,
		TotalPrice:      orderAggregate.Order.TotalPrice,
		DeliveredTime:   orderAggregate.Order.DeliveredTime,
		CancelReason:    orderAggregate.Order.CancelReason,
		DeliveryAddress: orderAggregate.Order.DeliveryAddress,
		Payment:         orderAggregate.Order.Payment,
	}
}

func OrderResponseFrom(projection *models.OrderProjection) dto.OrderResponseDto {
	return dto.OrderResponseDto{
		ID:              projection.ID,
		OrderID:         projection.OrderID,
		ShopItems:       ShopItemsResponseFromModels(projection.ShopItems),
		AccountEmail:    projection.AccountEmail,
		DeliveryAddress: projection.DeliveryAddress,
		CancelReason:    projection.CancelReason,
		TotalPrice:      projection.TotalPrice,
		DeliveredTime:   projection.DeliveredTime,
		Paid:            projection.Paid,
		Submitted:       projection.Submitted,
		Completed:       projection.Completed,
		Canceled:        projection.Canceled,
		Payment: dto.Payment{
			PaymentID: projection.Payment.PaymentID,
			Timestamp: projection.Payment.Timestamp,
		},
	}
}

func OrdersResponseFrom(projections []*models.OrderProjection) []dto.OrderResponseDto {
	orders := make([]dto.OrderResponseDto, 0, len(projections))
	for _, projection := range projections {
		orders = append(orders, OrderResponseFrom(projection))
	}
	return orders
}

func ShopItemsResponseFromModels(items []*models.ShopItem) []dto.ShopItem {
	shopItems := make([]dto.ShopItem, 0, len(items))
	for _, item := range items {
		shopItems = append(shopItems, dto.ShopItem{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			Quantity:    item.Quantity,
			Price:       item.Price,
		})
	}
	return shopItems
}
