package elastic

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/wassef911/astore/internal/delivery/aggregate"
	"github.com/wassef911/astore/internal/delivery/events"
	"github.com/wassef911/astore/internal/delivery/models"
	"github.com/wassef911/astore/internal/infrastructure/es"
	"github.com/wassef911/astore/internal/infrastructure/tracing"
)

func (o *elasticProjection) onOrderCreate(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.onOrderCreate")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var eventData events.OrderCreatedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	op := &models.OrderProjection{
		OrderID:      aggregate.GetOrderAggregateID(evt.AggregateID),
		ShopItems:    eventData.ShopItems,
		AccountEmail: eventData.AccountEmail,
		TotalPrice:   aggregate.GetShopItemsTotalPrice(eventData.ShopItems),
	}

	return o.elasticRepository.IndexOrder(ctx, op)
}

func (o *elasticProjection) onOrderPaid(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.onOrderPaid")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var payment models.Payment
	if err := evt.GetJsonData(&payment); err != nil {
		return errors.Wrap(err, "GetJsonData")
	}

	projection, err := o.elasticRepository.GetByID(ctx, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.Paid = true
	projection.Payment = payment

	return o.elasticRepository.UpdateOrder(ctx, projection)
}

func (o *elasticProjection) onSubmit(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.onSubmit")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	projection, err := o.elasticRepository.GetByID(ctx, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.Submitted = true

	return o.elasticRepository.UpdateOrder(ctx, projection)
}

func (o *elasticProjection) onShoppingCartUpdate(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.onShoppingCartUpdate")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var eventData events.ShoppingCartUpdatedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	projection, err := o.elasticRepository.GetByID(ctx, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.ShopItems = eventData.ShopItems
	projection.TotalPrice = aggregate.GetShopItemsTotalPrice(eventData.ShopItems)

	return o.elasticRepository.UpdateOrder(ctx, projection)
}

func (o *elasticProjection) onCancel(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.onCancel")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var eventData events.OrderCanceledEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	projection, err := o.elasticRepository.GetByID(ctx, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.Canceled = true
	projection.Completed = false
	projection.CancelReason = eventData.CancelReason

	return o.elasticRepository.UpdateOrder(ctx, projection)
}

func (o *elasticProjection) onComplete(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.onComplete")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var eventData events.OrderCompletedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	projection, err := o.elasticRepository.GetByID(ctx, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.Completed = true
	projection.DeliveredTime = eventData.DeliveryTimestamp

	return o.elasticRepository.UpdateOrder(ctx, projection)
}

func (o *elasticProjection) onDeliveryAddressChnaged(ctx context.Context, evt es.Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticProjection.onDeliveryAddressChnaged")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", evt.GetAggregateID()))

	var eventData events.OrderDeliveryAddressChangedEvent
	if err := evt.GetJsonData(&eventData); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "evt.GetJsonData")
	}

	projection, err := o.elasticRepository.GetByID(ctx, aggregate.GetOrderAggregateID(evt.AggregateID))
	if err != nil {
		return err
	}
	projection.DeliveryAddress = eventData.DeliveryAddress

	return o.elasticRepository.UpdateOrder(ctx, projection)

}
