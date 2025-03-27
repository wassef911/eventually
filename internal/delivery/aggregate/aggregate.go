package aggregate

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/wassef911/eventually/internal/delivery/events"
	"github.com/wassef911/eventually/internal/delivery/models"
	"github.com/wassef911/eventually/internal/infrastructure/es"
)

const (
	OrderAggregateType es.AggregateType = "order"
)

type InterfaceOrderAggregate interface {
	When(evt es.Event) error
	onOrderCreated(evt es.Event) error
	onOrderPaid(evt es.Event) error
	onOrderSubmitted(evt es.Event) error
	onOrderCompleted(evt es.Event) error
	onOrderCanceled(evt es.Event) error
	onShoppingCartUpdated(evt es.Event) error
	onChangeDeliveryAddress(evt es.Event) error
	CreateOrder(ctx context.Context, shopItems []*models.ShopItem, accountEmail, deliveryAddress string) error
	PayOrder(ctx context.Context, payment models.Payment) error
	SubmitOrder(ctx context.Context) error
	UpdateShoppingCart(ctx context.Context, shopItems []*models.ShopItem) error
	CancelOrder(ctx context.Context, cancelReason string) error
	CompleteOrder(ctx context.Context, deliveryTimestamp time.Time) error
	ChangeDeliveryAddress(ctx context.Context, deliveryAddress string) error
}

var _ InterfaceOrderAggregate = &OrderAggregate{}

type OrderAggregate struct {
	*es.AggregateBase
	Order *models.Order
}

func NewOrderAggregateWithID(id string) *OrderAggregate {
	if id == "" {
		return nil
	}

	aggregate := NewOrderAggregate()
	aggregate.SetID(id)
	aggregate.Order.ID = id
	return aggregate
}

func NewOrderAggregate() *OrderAggregate {
	orderAggregate := &OrderAggregate{Order: models.NewOrder()}
	base := es.NewAggregateBase(orderAggregate.When)
	base.SetType(OrderAggregateType)
	orderAggregate.AggregateBase = base
	return orderAggregate
}

func (a *OrderAggregate) When(evt es.Event) error {

	switch evt.GetEventType() {

	case events.OrderCreated:
		return a.onOrderCreated(evt)
	case events.OrderPaid:
		return a.onOrderPaid(evt)
	case events.OrderSubmitted:
		return a.onOrderSubmitted(evt)
	case events.OrderCompleted:
		return a.onOrderCompleted(evt)
	case events.OrderCanceled:
		return a.onOrderCanceled(evt)
	case events.ShoppingCartUpdated:
		return a.onShoppingCartUpdated(evt)
	case events.DeliveryAddressChanged:
		return a.onChangeDeliveryAddress(evt)

	default:
		return es.ErrInvalidEventType
	}
}

func (a *OrderAggregate) onOrderCreated(evt es.Event) error {
	var eventData events.OrderCreatedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		return errors.Wrap(err, "GetJsonData")
	}

	a.Order.AccountEmail = eventData.AccountEmail
	a.Order.ShopItems = eventData.ShopItems
	a.Order.TotalPrice = GetShopItemsTotalPrice(eventData.ShopItems)
	a.Order.DeliveryAddress = eventData.DeliveryAddress
	return nil
}

func (a *OrderAggregate) onOrderPaid(evt es.Event) error {
	var payment models.Payment
	if err := evt.GetJsonData(&payment); err != nil {
		return errors.Wrap(err, "GetJsonData")
	}

	a.Order.Paid = true
	a.Order.Payment = payment
	return nil
}

func (a *OrderAggregate) onOrderSubmitted(evt es.Event) error {
	a.Order.Submitted = true
	return nil
}

func (a *OrderAggregate) onOrderCompleted(evt es.Event) error {
	var eventData events.OrderCompletedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		return errors.Wrap(err, "GetJsonData")
	}

	a.Order.Completed = true
	a.Order.DeliveredTime = eventData.DeliveryTimestamp
	a.Order.Canceled = false
	return nil
}

func (a *OrderAggregate) onOrderCanceled(evt es.Event) error {
	var eventData events.OrderCanceledEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		return errors.Wrap(err, "GetJsonData")
	}

	a.Order.Canceled = true
	a.Order.Completed = false
	a.Order.CancelReason = eventData.CancelReason
	return nil
}

func (a *OrderAggregate) onShoppingCartUpdated(evt es.Event) error {
	var eventData events.ShoppingCartUpdatedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		return errors.Wrap(err, "GetJsonData")
	}

	a.Order.ShopItems = eventData.ShopItems
	a.Order.TotalPrice = GetShopItemsTotalPrice(eventData.ShopItems)
	return nil
}

func (a *OrderAggregate) onChangeDeliveryAddress(evt es.Event) error {
	var eventData events.OrderDeliveryAddressChangedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		return errors.Wrap(err, "GetJsonData")
	}

	a.Order.DeliveryAddress = eventData.DeliveryAddress
	return nil
}
