package utils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/wassef911/astore/internal/api/utils"
	"github.com/wassef911/astore/internal/delivery/models"
)

func TestOrderResponseFrom(t *testing.T) {
	projection := &models.OrderProjection{
		OrderID:         "order123",
		ShopItems:       []*models.ShopItem{{ID: "item1", Title: "Item 1", Description: "Description 1", Quantity: 2, Price: 10.0}},
		Paid:            true,
		Submitted:       true,
		Completed:       false,
		Canceled:        false,
		AccountEmail:    "test@example.com",
		TotalPrice:      20.0,
		DeliveredTime:   time.Now(),
		CancelReason:    "",
		DeliveryAddress: "123 Main St",
		Payment:         models.Payment{PaymentID: "pay123", Timestamp: time.Now()},
	}

	response := utils.OrderResponseFrom(projection)

	assert.Equal(t, projection.OrderID, response.OrderID)
	assert.Equal(t, projection.AccountEmail, response.AccountEmail)
	assert.Equal(t, projection.DeliveryAddress, response.DeliveryAddress)
	assert.Equal(t, projection.CancelReason, response.CancelReason)
	assert.Equal(t, projection.TotalPrice, response.TotalPrice)
	assert.Equal(t, projection.DeliveredTime, response.DeliveredTime)
	assert.Equal(t, projection.Paid, response.Paid)
	assert.Equal(t, projection.Submitted, response.Submitted)
	assert.Equal(t, projection.Completed, response.Completed)
	assert.Equal(t, projection.Canceled, response.Canceled)
	assert.Equal(t, projection.Payment.PaymentID, response.Payment.PaymentID)
	assert.Equal(t, projection.Payment.Timestamp, response.Payment.Timestamp)
}

func TestOrdersResponseFrom(t *testing.T) {
	projections := []*models.OrderProjection{
		{
			OrderID:         "order1",
			ShopItems:       []*models.ShopItem{{ID: "item1", Title: "Item 1", Description: "Description 1", Quantity: 2, Price: 10.0}},
			Paid:            true,
			Submitted:       true,
			Completed:       false,
			Canceled:        false,
			AccountEmail:    "test1@example.com",
			TotalPrice:      20.0,
			DeliveredTime:   time.Now(),
			CancelReason:    "",
			DeliveryAddress: "123 Main St",
			Payment:         models.Payment{PaymentID: "pay1", Timestamp: time.Now()},
		},
		{
			OrderID:         "order2",
			ShopItems:       []*models.ShopItem{{ID: "item2", Title: "Item 2", Description: "Description 2", Quantity: 1, Price: 15.0}},
			Paid:            false,
			Submitted:       true,
			Completed:       false,
			Canceled:        true,
			AccountEmail:    "test2@example.com",
			TotalPrice:      15.0,
			DeliveredTime:   time.Now(),
			CancelReason:    "Out of stock",
			DeliveryAddress: "456 Elm St",
			Payment:         models.Payment{PaymentID: "pay2", Timestamp: time.Now()},
		},
	}

	responses := utils.OrdersResponseFrom(projections)

	assert.Equal(t, len(projections), len(responses))
	for i, response := range responses {
		assert.Equal(t, projections[i].OrderID, response.OrderID)
		assert.Equal(t, projections[i].AccountEmail, response.AccountEmail)
		assert.Equal(t, projections[i].DeliveryAddress, response.DeliveryAddress)
		assert.Equal(t, projections[i].CancelReason, response.CancelReason)
		assert.Equal(t, projections[i].TotalPrice, response.TotalPrice)
		assert.Equal(t, projections[i].DeliveredTime, response.DeliveredTime)
		assert.Equal(t, projections[i].Paid, response.Paid)
		assert.Equal(t, projections[i].Submitted, response.Submitted)
		assert.Equal(t, projections[i].Completed, response.Completed)
		assert.Equal(t, projections[i].Canceled, response.Canceled)
		assert.Equal(t, projections[i].Payment.PaymentID, response.Payment.PaymentID)
		assert.Equal(t, projections[i].Payment.Timestamp, response.Payment.Timestamp)
	}
}

func TestShopItemsResponseFromModels(t *testing.T) {
	items := []*models.ShopItem{
		{ID: "item1", Title: "Item 1", Description: "Description 1", Quantity: 2, Price: 10.0},
		{ID: "item2", Title: "Item 2", Description: "Description 2", Quantity: 1, Price: 15.0},
	}

	shopItems := utils.ShopItemsResponseFromModels(items)

	assert.Equal(t, len(items), len(shopItems))
	for i, item := range items {
		assert.Equal(t, item.ID, shopItems[i].ID)
		assert.Equal(t, item.Title, shopItems[i].Title)
		assert.Equal(t, item.Description, shopItems[i].Description)
		assert.Equal(t, item.Quantity, shopItems[i].Quantity)
		assert.Equal(t, item.Price, shopItems[i].Price)
	}
}
