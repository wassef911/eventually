package service

import (
	"github.com/wassef911/astore/internal/delivery/commands"
	"github.com/wassef911/astore/internal/delivery/queries"
	"github.com/wassef911/astore/internal/delivery/repository"
	"github.com/wassef911/astore/internal/infrastructure/es/store"
	"github.com/wassef911/astore/pkg/config"
	"github.com/wassef911/astore/pkg/logger"
)

type OrderService struct {
	Commands *commands.OrderCommand
	Queries  *queries.OrderQueries
}

func New(
	log logger.Logger,
	cfg *config.Config,
	es store.AggregateStore,
	mongoRepo repository.OrderMongoRepository,
	elasticRepo repository.ElasticOrderRepository,
) *OrderService {

	createOrderHandler := commands.NewCreateOrderHandler(log, cfg, es)
	orderPaidHandler := commands.NewOrderPaidHandler(log, cfg, es)
	submitOrderHandler := commands.NewSubmitOrderHandler(log, cfg, es)
	updateOrderCmdHandler := commands.NewupdateShoppingCartCommandHandler(log, cfg, es)
	cancelOrderCommandHandler := commands.NewCancelOrderCommandHandler(log, cfg, es)
	deliveryOrderCommandHandler := commands.NewCompleteOrderCommandHandler(log, cfg, es)
	changeOrderDeliveryAddressCmdHandler := commands.NewchangeDeliveryAddressCommandHandler(log, cfg, es)

	getOrderByIDHandler := queries.NewGetOrderByIDHandler(log, cfg, es, mongoRepo)
	searchOrdersHandler := queries.NewSearchOrdersHandler(log, cfg, es, elasticRepo)

	orderCommands := commands.New(
		*createOrderHandler,
		*orderPaidHandler,
		*submitOrderHandler,
		*updateOrderCmdHandler,
		*cancelOrderCommandHandler,
		*deliveryOrderCommandHandler,
		*changeOrderDeliveryAddressCmdHandler,
	)
	orderQueries := queries.NewOrderQueries(getOrderByIDHandler, searchOrdersHandler)

	return &OrderService{Commands: orderCommands, Queries: orderQueries}
}
