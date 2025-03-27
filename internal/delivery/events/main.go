package events

import (
	"time"

	"github.com/wassef911/eventually/internal/delivery/models"
	"github.com/wassef911/eventually/internal/infrastructure/es"
)

const (
	OrderCreated           = "ORDER_CREATED"
	OrderPaid              = "ORDER_PAID"
	OrderSubmitted         = "ORDER_SUBMITTED"
	OrderCompleted         = "ORDER_COMPLETED"
	OrderCanceled          = "ORDER_CANCELED"
	ShoppingCartUpdated    = "SHOPPING_CART_UPDATED"
	DeliveryAddressChanged = "DELIVERY_ADDRESS_CHANGED"
)

type OrderCreatedEvent struct {
	ShopItems       []*models.ShopItem `json:"shopItems" bson:"shopItems,omitempty"`
	AccountEmail    string             `json:"accountEmail" bson:"accountEmail,omitempty"`
	DeliveryAddress string             `json:"deliveryAddress" bson:"deliveryAddress,omitempty"`
}

func NewOrderCreatedEvent(aggregate es.Aggregate, shopItems []*models.ShopItem, accountEmail, deliveryAddress string) (es.Event, error) {
	eventData := OrderCreatedEvent{
		ShopItems:       shopItems,
		AccountEmail:    accountEmail,
		DeliveryAddress: deliveryAddress,
	}
	event := es.NewBaseEvent(aggregate, OrderCreated)
	if err := event.SetJsonData(&eventData); err != nil {
		return es.Event{}, err
	}
	return event, nil
}

func NewOrderPaidEvent(aggregate es.Aggregate, payment *models.Payment) (es.Event, error) {
	event := es.NewBaseEvent(aggregate, OrderPaid)
	if err := event.SetJsonData(&payment); err != nil {
		return es.Event{}, err
	}
	return event, nil
}

func NewSubmitOrderEvent(aggregate es.Aggregate) (es.Event, error) {
	return es.NewBaseEvent(aggregate, OrderSubmitted), nil
}

type ShoppingCartUpdatedEvent struct {
	ShopItems []*models.ShopItem `json:"shopItems" bson:"shopItems,omitempty"`
}

func NewShoppingCartUpdatedEvent(aggregate es.Aggregate, shopItems []*models.ShopItem) (es.Event, error) {
	eventData := ShoppingCartUpdatedEvent{ShopItems: shopItems}
	event := es.NewBaseEvent(aggregate, ShoppingCartUpdated)
	if err := event.SetJsonData(&eventData); err != nil {
		return es.Event{}, err
	}
	return event, nil
}

type OrderDeliveryAddressChangedEvent struct {
	DeliveryAddress string `json:"deliveryAddress" bson:"deliveryAddress,omitempty"`
}

func NewDeliveryAddressChangedEvent(aggregate es.Aggregate, deliveryAddress string) (es.Event, error) {
	eventData := OrderDeliveryAddressChangedEvent{DeliveryAddress: deliveryAddress}
	event := es.NewBaseEvent(aggregate, DeliveryAddressChanged)
	if err := event.SetJsonData(&eventData); err != nil {
		return es.Event{}, err
	}
	return event, nil
}

type OrderCanceledEvent struct {
	CancelReason string `json:"cancelReason"`
}

func NewOrderCanceledEvent(aggregate es.Aggregate, cancelReason string) (es.Event, error) {
	eventData := OrderCanceledEvent{CancelReason: cancelReason}
	event := es.NewBaseEvent(aggregate, OrderCanceled)
	err := event.SetJsonData(&eventData)
	if err != nil {
		return es.Event{}, err
	}
	return event, nil
}

type OrderCompletedEvent struct {
	DeliveryTimestamp time.Time `json:"deliveryTimestamp"`
}

func NewOrderCompletedEvent(aggregate es.Aggregate, deliveryTimestamp time.Time) (es.Event, error) {
	eventData := OrderCompletedEvent{DeliveryTimestamp: deliveryTimestamp}
	event := es.NewBaseEvent(aggregate, OrderCompleted)
	err := event.SetJsonData(&eventData)
	if err != nil {
		return es.Event{}, err
	}
	return event, nil
}
