package commands

import (
	"context"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/wassef911/astore/internal/delivery/aggregate"
	"github.com/wassef911/astore/internal/infrastructure/es/store"
	"github.com/wassef911/astore/pkg/config"
	"github.com/wassef911/astore/pkg/logger"
)

// all handlers implement commandHandler
var _ commandHandler[*CancelOrderCommand] = &cancelOrderCommandHandler{}
var _ commandHandler[*ChangeDeliveryAddressCommand] = &changeDeliveryAddressCommandHandler{}
var _ commandHandler[*CompleteOrderCommand] = &completeOrderCommandHandler{}
var _ commandHandler[*CreateOrderCommand] = &createOrderHandler{}
var _ commandHandler[*PayOrderCommand] = &payOrderCommandHandler{}
var _ commandHandler[*SubmitOrderCommand] = &submitOrderCommandHandler{}
var _ commandHandler[*UpdateShoppingCartCommand] = &updateShoppingCartCommandHandler{}

type cancelOrderCommandHandler struct {
	baseCommandHandler
}

func NewCancelOrderCommandHandler(log logger.Logger, cfg *config.Config, es store.AggregateStore) *cancelOrderCommandHandler {
	return &cancelOrderCommandHandler{baseCommandHandler{log: log, cfg: cfg, es: es}}
}

func (c *cancelOrderCommandHandler) Handle(ctx context.Context, command *CancelOrderCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cancelOrderCommandHandler.Handle")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.CancelOrder(ctx, command.CancelReason); err != nil {
		return err
	}

	return c.es.Save(ctx, order)
}

type changeDeliveryAddressCommandHandler struct {
	baseCommandHandler
}

func NewchangeDeliveryAddressCommandHandler(log logger.Logger, cfg *config.Config, es store.AggregateStore) *changeDeliveryAddressCommandHandler {
	return &changeDeliveryAddressCommandHandler{baseCommandHandler{log: log, cfg: cfg, es: es}}
}

func (c *changeDeliveryAddressCommandHandler) Handle(ctx context.Context, command *ChangeDeliveryAddressCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "changeDeliveryAddressCommandHandler.Handle")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.ChangeDeliveryAddress(ctx, command.DeliveryAddress); err != nil {
		return err
	}

	return c.es.Save(ctx, order)
}

type completeOrderCommandHandler struct {
	baseCommandHandler
}

func NewCompleteOrderCommandHandler(log logger.Logger, cfg *config.Config, es store.AggregateStore) *completeOrderCommandHandler {
	return &completeOrderCommandHandler{baseCommandHandler{log: log, cfg: cfg, es: es}}
}

func (c *completeOrderCommandHandler) Handle(ctx context.Context, command *CompleteOrderCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "completeOrderCommandHandler.Handle")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.CompleteOrder(ctx, command.DeliveryTimestamp); err != nil {
		return err
	}

	return c.es.Save(ctx, order)
}

type createOrderHandler struct {
	baseCommandHandler
}

func NewCreateOrderHandler(log logger.Logger, cfg *config.Config, es store.AggregateStore) *createOrderHandler {
	return &createOrderHandler{baseCommandHandler{log: log, cfg: cfg, es: es}}
}

func (c *createOrderHandler) Handle(ctx context.Context, command *CreateOrderCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "createOrderHandler.Handle")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", command.GetAggregateID()))

	order := aggregate.NewOrderAggregateWithID(command.AggregateID)
	err := c.es.Exists(ctx, order.GetID())
	if err != nil && !errors.Is(err, esdb.ErrStreamNotFound) {
		return err
	}

	if err := order.CreateOrder(ctx, command.ShopItems, command.AccountEmail, command.DeliveryAddress); err != nil {
		return err
	}

	span.LogFields(log.String("order", order.String()))
	return c.es.Save(ctx, order)
}

type payOrderCommandHandler struct {
	baseCommandHandler
}

func NewOrderPaidHandler(log logger.Logger, cfg *config.Config, es store.AggregateStore) *payOrderCommandHandler {
	return &payOrderCommandHandler{baseCommandHandler{log: log, cfg: cfg, es: es}}
}

func (c *payOrderCommandHandler) Handle(ctx context.Context, command *PayOrderCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "payOrderCommandHandler.Handle")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.PayOrder(ctx, command.Payment); err != nil {
		return err
	}

	return c.es.Save(ctx, order)
}

type submitOrderCommandHandler struct {
	baseCommandHandler
}

func NewSubmitOrderHandler(log logger.Logger, cfg *config.Config, es store.AggregateStore) *submitOrderCommandHandler {
	return &submitOrderCommandHandler{baseCommandHandler{log: log, cfg: cfg, es: es}}
}

func (c *submitOrderCommandHandler) Handle(ctx context.Context, command *SubmitOrderCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "submitOrderHandler.Handle")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.SubmitOrder(ctx); err != nil {
		return err
	}

	return c.es.Save(ctx, order)
}

type updateShoppingCartCommandHandler struct {
	baseCommandHandler
}

func NewupdateShoppingCartCommandHandler(log logger.Logger, cfg *config.Config, es store.AggregateStore) *updateShoppingCartCommandHandler {
	return &updateShoppingCartCommandHandler{baseCommandHandler{log: log, cfg: cfg, es: es}}
}

func (c *updateShoppingCartCommandHandler) Handle(ctx context.Context, command *UpdateShoppingCartCommand) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "updateShoppingCartCommandHandler.Handle")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", command.GetAggregateID()))

	order, err := aggregate.LoadOrderAggregate(ctx, c.es, command.GetAggregateID())
	if err != nil {
		return err
	}

	if err := order.UpdateShoppingCart(ctx, command.ShopItems); err != nil {
		return err
	}

	return c.es.Save(ctx, order)
}
