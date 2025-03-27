package repository

import (
	"context"

	"github.com/wassef911/eventually/internal/api/dto"
	"github.com/wassef911/eventually/internal/api/utils"
	"github.com/wassef911/eventually/internal/delivery/models"
)

type OrderMongoRepository interface {
	Insert(ctx context.Context, order *models.OrderProjection) (string, error)
	GetByID(ctx context.Context, orderID string) (*models.OrderProjection, error)
	UpdateOrder(ctx context.Context, order *models.OrderProjection) error

	UpdateCancel(ctx context.Context, order *models.OrderProjection) error
	UpdatePayment(ctx context.Context, order *models.OrderProjection) error
	Complete(ctx context.Context, order *models.OrderProjection) error
	UpdateDeliveryAddress(ctx context.Context, order *models.OrderProjection) error
	UpdateSubmit(ctx context.Context, order *models.OrderProjection) error
}

type ElasticOrderRepository interface {
	IndexOrder(ctx context.Context, order *models.OrderProjection) error
	GetByID(ctx context.Context, orderID string) (*models.OrderProjection, error)
	UpdateOrder(ctx context.Context, order *models.OrderProjection) error
	Search(ctx context.Context, text string, pq *utils.Pagination) (*dto.OrderSearchResponseDto, error)
}
