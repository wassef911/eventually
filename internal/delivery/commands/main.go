package commands

type OrderCommand struct {
	CreateOrder                createOrderHandler
	OrderPaid                  payOrderCommandHandler
	SubmitOrder                submitOrderCommandHandler
	UpdateOrder                updateShoppingCartCommandHandler
	CancelOrder                cancelOrderCommandHandler
	CompleteOrder              completeOrderCommandHandler
	ChangeOrderDeliveryAddress changeDeliveryAddressCommandHandler
}

func New(
	createOrder createOrderHandler,
	orderPaid payOrderCommandHandler,
	submitOrder submitOrderCommandHandler,
	updateOrder updateShoppingCartCommandHandler,
	cancelOrder cancelOrderCommandHandler,
	completeOrder completeOrderCommandHandler,
	changeOrderDeliveryAddress changeDeliveryAddressCommandHandler,
) *OrderCommand {
	return &OrderCommand{
		CreateOrder:                createOrder,
		OrderPaid:                  orderPaid,
		SubmitOrder:                submitOrder,
		UpdateOrder:                updateOrder,
		CancelOrder:                cancelOrder,
		CompleteOrder:              completeOrder,
		ChangeOrderDeliveryAddress: changeOrderDeliveryAddress,
	}
}
