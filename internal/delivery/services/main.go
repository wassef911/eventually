package service

import (
	"github.com/wassef911/eventually/internal/delivery/commands"
	"github.com/wassef911/eventually/internal/delivery/queries"
	"github.com/wassef911/eventually/internal/delivery/repository"
	"github.com/wassef911/eventually/internal/infrastructure/es/store"
	"github.com/wassef911/eventually/pkg/config"
	"github.com/wassef911/eventually/pkg/logger"
)

type OrderService struct {
	Commands *commands.OrderCommand
	Queries  *queries.OrderQueries
}

func New(
	log logger.Logger,
	config *config.Config,
	es store.AggregateStore,
	mongoRepo repository.OrderMongoRepository,
	elasticRepo repository.ElasticOrderRepository,
) *OrderService {

	createOrderHandler := commands.NewCreateOrderHandler(log, config, es)
	orderPaidHandler := commands.NewOrderPaidHandler(log, config, es)
	submitOrderHandler := commands.NewSubmitOrderHandler(log, config, es)
	updateOrderCmdHandler := commands.NewupdateShoppingCartCommandHandler(log, config, es)
	cancelOrderCommandHandler := commands.NewCancelOrderCommandHandler(log, config, es)
	deliveryOrderCommandHandler := commands.NewCompleteOrderCommandHandler(log, config, es)
	changeOrderDeliveryAddressCmdHandler := commands.NewchangeDeliveryAddressCommandHandler(log, config, es)

	getOrderByIDHandler := queries.NewGetOrderByIDHandler(log, config, es, mongoRepo)
	searchOrdersHandler := queries.NewSearchOrdersHandler(log, config, es, elasticRepo)

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
