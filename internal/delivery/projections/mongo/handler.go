package mongo

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/wassef911/eventually/internal/delivery/aggregate"
	"github.com/wassef911/eventually/internal/delivery/events"
	"github.com/wassef911/eventually/internal/delivery/models"
	"github.com/wassef911/eventually/internal/infrastructure/es"
	"github.com/wassef911/eventually/internal/infrastructure/tracing"
)

func (o *mongoProjection) onOrderCreate(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mongoProjection.onOrderCreate")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var eventData events.OrderCreatedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "evt.GetJsonData")
	}
	span.LogFields(log.String("AccountEmail", eventData.AccountEmail))

	op := &models.OrderProjection{
		OrderID:         aggregate.GetOrderAggregateID(evt.AggregateID),
		ShopItems:       eventData.ShopItems,
		AccountEmail:    eventData.AccountEmail,
		TotalPrice:      aggregate.GetShopItemsTotalPrice(eventData.ShopItems),
		DeliveryAddress: eventData.DeliveryAddress,
	}

	_, err := o.mongoRepo.Insert(ctx, op)
	if err != nil {
		return err
	}

	return nil
}

func (o *mongoProjection) onOrderPaid(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mongoProjection.onOrderPaid")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var payment models.Payment
	if err := evt.GetJsonData(&payment); err != nil {
		return errors.Wrap(err, "GetJsonData")
	}

	op := &models.OrderProjection{OrderID: aggregate.GetOrderAggregateID(evt.AggregateID), Paid: true, Payment: payment}
	return o.mongoRepo.UpdatePayment(ctx, op)
}

func (o *mongoProjection) onSubmit(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mongoProjection.onSubmit")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	op := &models.OrderProjection{OrderID: aggregate.GetOrderAggregateID(evt.AggregateID), Submitted: true}
	return o.mongoRepo.UpdateSubmit(ctx, op)
}

func (o *mongoProjection) onShoppingCartUpdate(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mongoProjection.onShoppingCartUpdate")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var eventData events.ShoppingCartUpdatedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	op := &models.OrderProjection{OrderID: aggregate.GetOrderAggregateID(evt.AggregateID), ShopItems: eventData.ShopItems}
	op.TotalPrice = aggregate.GetShopItemsTotalPrice(eventData.ShopItems)
	return o.mongoRepo.UpdateOrder(ctx, op)
}

func (o *mongoProjection) onCancel(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mongoProjection.onCancel")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var eventData events.OrderCanceledEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	op := &models.OrderProjection{
		OrderID:      aggregate.GetOrderAggregateID(evt.AggregateID),
		Canceled:     true,
		Completed:    false,
		CancelReason: eventData.CancelReason,
	}
	return o.mongoRepo.UpdateCancel(ctx, op)
}

func (o *mongoProjection) onCompleted(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mongoProjection.onCompleted")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var eventData events.OrderCompletedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	op := &models.OrderProjection{
		OrderID:       aggregate.GetOrderAggregateID(evt.AggregateID),
		Canceled:      false,
		Completed:     true,
		DeliveredTime: eventData.DeliveryTimestamp,
	}
	return o.mongoRepo.Complete(ctx, op)
}

func (o *mongoProjection) onDeliveryAddressChanged(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "mongoProjection.onDeliveryAddressChanged")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var eventData events.OrderDeliveryAddressChangedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	op := &models.OrderProjection{
		OrderID:         aggregate.GetOrderAggregateID(evt.AggregateID),
		DeliveryAddress: eventData.DeliveryAddress,
	}
	return o.mongoRepo.UpdateDeliveryAddress(ctx, op)
}
