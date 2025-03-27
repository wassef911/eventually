package queries

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wassef911/eventually/internal/api/dto"
	"github.com/wassef911/eventually/internal/api/utils"
	"github.com/wassef911/eventually/internal/delivery/aggregate"
	"github.com/wassef911/eventually/internal/delivery/models"
	"github.com/wassef911/eventually/internal/delivery/repository"
	"github.com/wassef911/eventually/internal/infrastructure/es/store"
	"github.com/wassef911/eventually/pkg/config"
	"github.com/wassef911/eventually/pkg/logger"
)

type SearchOrdersQueryHandler interface {
	Handle(ctx context.Context, command *SearchOrdersQuery) (*dto.OrderSearchResponseDto, error)
}

type searchOrdersHandler struct {
	log               logger.Logger
	cfg               *config.Config
	es                store.AggregateStore
	elasticRepository repository.ElasticOrderRepository
}

func NewSearchOrdersHandler(log logger.Logger, cfg *config.Config, es store.AggregateStore, elasticRepository repository.ElasticOrderRepository) *searchOrdersHandler {
	return &searchOrdersHandler{log: log, cfg: cfg, es: es, elasticRepository: elasticRepository}
}

func (s *searchOrdersHandler) Handle(ctx context.Context, query *SearchOrdersQuery) (*dto.OrderSearchResponseDto, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "searchOrdersHandler.Handle")
	defer span.Finish()
	span.LogFields(log.String("SearchText", query.SearchText))

	return s.elasticRepository.Search(ctx, query.SearchText, query.Pq)
}

type GetOrderByIDQueryHandler interface {
	Handle(ctx context.Context, command *GetOrderByIDQuery) (*models.OrderProjection, error)
}

type getOrderByIDHandler struct {
	log       logger.Logger
	cfg       *config.Config
	es        store.AggregateStore
	mongoRepo repository.OrderMongoRepository
}

func NewGetOrderByIDHandler(log logger.Logger, cfg *config.Config, es store.AggregateStore, mongoRepo repository.OrderMongoRepository) *getOrderByIDHandler {
	return &getOrderByIDHandler{log: log, cfg: cfg, es: es, mongoRepo: mongoRepo}
}

func (q *getOrderByIDHandler) Handle(ctx context.Context, query *GetOrderByIDQuery) (*models.OrderProjection, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "getOrderByIDHandler.Handle")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", query.ID))

	orderProjection, err := q.mongoRepo.GetByID(ctx, query.ID)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	if orderProjection != nil {
		return orderProjection, nil
	}

	order := aggregate.NewOrderAggregateWithID(query.ID)
	if err := q.es.Load(ctx, order); err != nil {
		return nil, err
	}

	if aggregate.IsAggregateNotFound(order) {
		return nil, aggregate.ErrOrderNotFound
	}

	orderProjection = utils.OrderProjectionFrom(order)

	_, err = q.mongoRepo.Insert(ctx, orderProjection)
	if err != nil {
		return nil, err
	}

	return orderProjection, nil
}
