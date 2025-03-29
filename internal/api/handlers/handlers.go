package handlers

import (
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"

	"github.com/wassef911/eventually/internal/api/constants"
	"github.com/wassef911/eventually/internal/api/dto"
	api "github.com/wassef911/eventually/internal/api/middlewares"
	"github.com/wassef911/eventually/internal/api/utils"
	"github.com/wassef911/eventually/internal/delivery/commands"
	"github.com/wassef911/eventually/internal/delivery/models"
	"github.com/wassef911/eventually/internal/delivery/queries"
	service "github.com/wassef911/eventually/internal/delivery/services"
	"github.com/wassef911/eventually/internal/infrastructure/tracing"
	"github.com/wassef911/eventually/pkg/config"
	"github.com/wassef911/eventually/pkg/errors"
	"github.com/wassef911/eventually/pkg/logger"
)

type OrderHandlersI interface {
	CreateOrder() echo.HandlerFunc
	PayOrder() echo.HandlerFunc
	SubmitOrder() echo.HandlerFunc
	UpdateShoppingCart() echo.HandlerFunc
	MapRoutes()
	GetOrderByID() echo.HandlerFunc
	Search() echo.HandlerFunc
}

var _ OrderHandlersI = &orderHandlers{}

type orderHandlers struct {
	group *echo.Group
	log   logger.Logger
	mw    api.MiddlewareManager
	cfg   *config.Config
	v     *validator.Validate
	os    *service.OrderService
}

func NewOrderHandlers(
	group *echo.Group,
	log logger.Logger,
	mw api.MiddlewareManager,
	cfg *config.Config,
	v *validator.Validate,
	os *service.OrderService,
) *orderHandlers {
	return &orderHandlers{group: group, log: log, mw: mw, cfg: cfg, v: v, os: os}
}

func (h *orderHandlers) MapRoutes() {
	h.group.POST("", h.CreateOrder())
	h.group.PUT("/pay/:id", h.PayOrder())
	h.group.PUT("/submit/:id", h.SubmitOrder())
	h.group.PUT("/cart/:id", h.UpdateShoppingCart())
	h.group.POST("/cancel/:id", h.CancelOrder())
	h.group.POST("/complete/:id", h.CompleteOrder())
	h.group.PUT("/address/:id", h.ChangeDeliveryAddress())

	h.group.GET("/:id", h.GetOrderByID())
	h.group.GET("/search", h.Search())
}

// CreateOrder
// @Tags Orders
// @Summary Create order
// @Description Create new order
// @Param order body dto.CreateOrderReqDto true "create order"
// @Accept json
// @Produce json
// @Success 201 {string} id ""
// @Router /orders [post]
func (h *orderHandlers) CreateOrder() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "orderHandlers.CreateOrder")
		defer span.Finish()

		var reqDto dto.CreateOrderReqDto
		if err := c.Bind(&reqDto); err != nil {
			return err
		}

		if err := h.v.StructCtx(ctx, reqDto); err != nil {
			return err
		}

		id := uuid.NewV4().String()
		command := commands.NewCreateOrderCommand(id, reqDto.ShopItems, reqDto.AccountEmail, reqDto.DeliveryAddress)
		err := h.os.Commands.CreateOrder.Handle(ctx, command)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, id)
	}
}

// PayOrder
// @Tags Orders
// @Summary Pay order
// @Description Pay existing order
// @Accept json
// @Produce json
// @Param order body dto.Payment true "create order"
// @Param id path string true "Order ID"
// @Success 200 {string} id ""
// @Router /orders/pay/{id} [put]
func (h *orderHandlers) PayOrder() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "orderHandlers.PayOrder")
		defer span.Finish()

		orderID, err := uuid.FromString(c.Param(constants.ID))
		if err != nil {
			return err
		}

		var payment dto.Payment
		if err := c.Bind(&payment); err != nil {
			return err
		}

		command := commands.NewPayOrderCommand(models.Payment{PaymentID: payment.PaymentID, Timestamp: payment.Timestamp}, orderID.String())
		if err := h.v.StructCtx(ctx, command); err != nil {
			return err
		}

		err = h.os.Commands.OrderPaid.Handle(ctx, command)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, orderID.String())
	}
}

// SubmitOrder
// @Tags Orders
// @Summary Submit order
// @Description Submit existing order
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {string} id ""
// @Router /orders/submit/{id} [put]
func (h *orderHandlers) SubmitOrder() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "orderHandlers.SubmitOrder")
		defer span.Finish()

		orderID, err := uuid.FromString(c.Param(constants.ID))
		if err != nil {
			return err
		}

		command := commands.NewSubmitOrderCommand(orderID.String())
		if err := h.v.StructCtx(ctx, command); err != nil {
			return err
		}

		err = h.os.Commands.SubmitOrder.Handle(ctx, command)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, orderID.String())
	}
}

// CancelOrder
// @Tags Orders
// @Summary Cancel order
// @Description Cancel existing order
// @Accept json
// @Produce json
// @Param order body dto.CancelOrderReqDto true "cancel order reason"
// @Param id path string true "Order ID"
// @Success 200 {string} id ""
// @Router /orders/cancel/{id} [post]
func (h *orderHandlers) CancelOrder() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "orderHandlers.CancelOrder")
		defer span.Finish()

		orderID, err := uuid.FromString(c.Param(constants.ID))
		if err != nil {
			return err
		}

		var data dto.CancelOrderReqDto
		if err := c.Bind(&data); err != nil {
			return err
		}

		command := commands.NewCancelOrderCommand(orderID.String(), data.CancelReason)
		if err := h.v.StructCtx(ctx, command); err != nil {
			return err
		}

		err = h.os.Commands.CancelOrder.Handle(ctx, command)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, orderID.String())
	}
}

// CompleteOrder
// @Tags Orders
// @Summary Complete order
// @Description Complete existing order
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {string} id ""
// @Router /orders/complete/{id} [post]
func (h *orderHandlers) CompleteOrder() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "orderHandlers.CompleteOrder")
		defer span.Finish()

		orderID, err := uuid.FromString(c.Param(constants.ID))
		if err != nil {
			return err
		}

		command := commands.NewCompleteOrderCommand(orderID.String(), time.Now())
		if err := h.v.StructCtx(ctx, command); err != nil {
			return err
		}

		err = h.os.Commands.CompleteOrder.Handle(ctx, command)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, orderID.String())
	}
}

// ChangeDeliveryAddress
// @Tags Orders
// @Summary Change delivery address order
// @Description Deliver existing order
// @Accept json
// @Produce json
// @Param order body dto.ChangeDeliveryAddressReqDto true "change delivery address"
// @Param id path string true "Order ID"
// @Success 200 {string} id ""
// @Router /orders/address/{id} [put]
func (h *orderHandlers) ChangeDeliveryAddress() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "orderHandlers.ChangeDeliveryAddress")
		defer span.Finish()

		param := c.Param(constants.ID)
		orderID, err := uuid.FromString(param)
		if err != nil {
			return err
		}

		var data dto.ChangeDeliveryAddressReqDto
		if err := c.Bind(&data); err != nil {
			tracing.TraceErr(span, err)
			return errors.ErrorCtxResponse(c, err, h.cfg.Logger.Debug)
		}

		command := commands.NewChangeDeliveryAddressCommand(orderID.String(), data.DeliveryAddress)
		if err := h.v.StructCtx(ctx, command); err != nil {
			return err
		}

		err = h.os.Commands.ChangeOrderDeliveryAddress.Handle(ctx, command)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, orderID.String())
	}
}

// UpdateShoppingCart
// @Tags Orders
// @Summary Update order shopping cart
// @Description Update existing order shopping cart
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param order body dto.UpdateShoppingItemsReqDto true "update order"
// @Success 200 {string} id ""
// @Router /orders/cart/{id} [put]
func (h *orderHandlers) UpdateShoppingCart() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "orderHandlers.UpdateShoppingCart")
		defer span.Finish()

		orderID, err := uuid.FromString(c.Param(constants.ID))
		if err != nil {
			return err
		}

		var reqDto dto.UpdateShoppingItemsReqDto
		if err := c.Bind(&reqDto); err != nil {
			return err
		}

		if err := h.v.StructCtx(ctx, reqDto); err != nil {
			return err
		}

		command := commands.NewUpdateShoppingCartCommand(orderID.String(), reqDto.ShopItems)
		err = h.os.Commands.UpdateOrder.Handle(ctx, command)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, orderID.String())
	}
}

// GetOrderByID
// @Tags Orders
// @Summary Get order
// @Description Get order by id
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} dto.OrderResponseDto
// @Router /orders/{id} [get]
func (h *orderHandlers) GetOrderByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "orderHandlers.GetOrderByID")
		defer span.Finish()

		param := c.Param(constants.ID)
		orderID, err := uuid.FromString(param)
		if err != nil {
			return err
		}

		query := queries.NewGetOrderByIDQuery(orderID.String())
		if err := h.v.StructCtx(ctx, query); err != nil {
			return err
		}

		orderProjection, err := h.os.Queries.GetOrderByID.Handle(ctx, query)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, utils.OrderResponseFrom(orderProjection))
	}
}

// Search
// @Tags Orders
// @Summary Search orders
// @Description Full text search by title and description
// @Accept json
// @Produce json
// @Param search query string false "search text"
// @Param page query string false "page number"
// @Param size query string false "number of elements"
// @Success 200 {object} dto.OrderSearchResponseDto
// @Router /orders/search [get]
func (h *orderHandlers) Search() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "orderHandlers.Search")
		defer span.Finish()

		pq := utils.NewPaginationFromQueryParams(c.QueryParam(constants.Size), c.QueryParam(constants.Page))

		query := queries.NewSearchOrdersQuery(c.QueryParam(constants.Search), pq)
		if err := h.v.StructCtx(ctx, query); err != nil {
			return err
		}

		searchRes, err := h.os.Queries.SearchOrders.Handle(ctx, query)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, searchRes)
	}
}
