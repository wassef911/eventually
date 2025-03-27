package aggregate

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/wassef911/astore/internal/delivery/events"
	"github.com/wassef911/astore/internal/delivery/models"
	"github.com/wassef911/astore/internal/infrastructure/tracing"
)

func (a *OrderAggregate) CreateOrder(ctx context.Context, shopItems []*models.ShopItem, accountEmail, deliveryAddress string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "OrderAggregate.CreateOrder")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	if shopItems == nil {
		return ErrOrderShopItemsIsRequired
	}
	if deliveryAddress == "" {
		return ErrInvalidDeliveryAddress
	}

	event, err := events.NewOrderCreatedEvent(a, shopItems, accountEmail, deliveryAddress)
	if err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "NewOrderCreatedEvent")
	}

	if err := event.SetMetadata(tracing.ExtractTextMapCarrier(span.Context())); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}

func (a *OrderAggregate) PayOrder(ctx context.Context, payment models.Payment) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "OrderAggregate.PayOrder")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	if a.Order.Canceled {
		return ErrOrderAlreadyCancelled
	}
	if a.Order.Paid {
		return ErrAlreadyPaid
	}
	if a.Order.Submitted {
		return ErrAlreadySubmitted
	}

	event, err := events.NewOrderPaidEvent(a, &payment)
	if err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "NewOrderPaidEvent")
	}

	if err := event.SetMetadata(tracing.ExtractTextMapCarrier(span.Context())); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}

func (a *OrderAggregate) SubmitOrder(ctx context.Context) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "OrderAggregate.SubmitOrder")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	if a.Order.Canceled {
		return ErrOrderAlreadyCancelled
	}
	if !a.Order.Paid {
		return ErrOrderNotPaid
	}
	if a.Order.Submitted {
		return ErrAlreadySubmitted
	}

	submitOrderEvent, err := events.NewSubmitOrderEvent(a)
	if err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "NewSubmitOrderEvent")
	}

	if err := submitOrderEvent.SetMetadata(tracing.ExtractTextMapCarrier(span.Context())); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(submitOrderEvent)
}

func (a *OrderAggregate) UpdateShoppingCart(ctx context.Context, shopItems []*models.ShopItem) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "OrderAggregate.UpdateShoppingCart")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	if a.Order.Canceled {
		return ErrOrderAlreadyCancelled
	}
	if a.Order.Submitted {
		return ErrAlreadySubmitted
	}

	orderUpdatedEvent, err := events.NewShoppingCartUpdatedEvent(a, shopItems)
	if err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "NewShoppingCartUpdatedEvent")
	}

	if err := orderUpdatedEvent.SetMetadata(tracing.ExtractTextMapCarrier(span.Context())); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(orderUpdatedEvent)
}

func (a *OrderAggregate) CancelOrder(ctx context.Context, cancelReason string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "OrderAggregate.CancelOrder")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	if a.Order.Completed {
		return ErrOrderAlreadyCompleted
	}
	if cancelReason == "" {
		return ErrCancelReasonRequired
	}

	event, err := events.NewOrderCanceledEvent(a, cancelReason)
	if err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "NewOrderCanceledEvent")
	}

	if err := event.SetMetadata(tracing.ExtractTextMapCarrier(span.Context())); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}

func (a *OrderAggregate) CompleteOrder(ctx context.Context, deliveryTimestamp time.Time) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "OrderAggregate.CompleteOrder")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	if a.Order.Completed {
		return ErrOrderAlreadyCompleted
	}
	if a.Order.Canceled {
		return ErrOrderAlreadyCanceled
	}
	if !a.Order.Paid {
		return ErrOrderMustBePaidBeforeDelivered
	}

	event, err := events.NewOrderCompletedEvent(a, deliveryTimestamp)
	if err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "NewOrderCompletedEvent")
	}

	if err := event.SetMetadata(tracing.ExtractTextMapCarrier(span.Context())); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}

func (a *OrderAggregate) ChangeDeliveryAddress(ctx context.Context, deliveryAddress string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "OrderAggregate.ChangeDeliveryAddress")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	if a.Order.Completed {
		return ErrOrderAlreadyCompleted
	}

	event, err := events.NewDeliveryAddressChangedEvent(a, deliveryAddress)
	if err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "NewDeliveryAddressChangedEvent")
	}

	if err := event.SetMetadata(tracing.ExtractTextMapCarrier(span.Context())); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "SetMetadata")
	}

	return a.Apply(event)
}
