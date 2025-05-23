package repository

import (
	"context"
	"encoding/json"

	v7 "github.com/olivere/elastic/v7"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/wassef911/eventually/internal/api/dto"
	"github.com/wassef911/eventually/internal/api/utils"
	"github.com/wassef911/eventually/internal/delivery/models"
	"github.com/wassef911/eventually/internal/infrastructure/tracing"
	"github.com/wassef911/eventually/pkg/config"
	"github.com/wassef911/eventually/pkg/logger"
)

const (
	shopItemTitle            = "shopItems.title"
	shopItemDescription      = "shopItems.description"
	minimumNumberShouldMatch = 1
)

type ElasticRepository struct {
	log           logger.Logger
	config        *config.Config
	elasticClient *v7.Client
}

func NewElasticRepository(log logger.Logger, config *config.Config, elasticClient *v7.Client) *ElasticRepository {
	return &ElasticRepository{log: log, config: config, elasticClient: elasticClient}
}

func (e ElasticRepository) IndexOrder(ctx context.Context, order *models.OrderProjection) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepository.IndexOrder")
	defer span.Finish()
	span.LogFields(log.String("OrderID", order.OrderID))

	_, err := e.elasticClient.Index().Index(e.config.ElasticIndexes.Orders).BodyJson(order).Id(order.OrderID).Do(ctx)
	if err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "elasticClient.Index")
	}

	return nil
}

func (e ElasticRepository) GetByID(ctx context.Context, orderID string) (*models.OrderProjection, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepository.GetByID")
	defer span.Finish()
	span.LogFields(log.String("OrderID", orderID))

	result, err := e.elasticClient.Get().Index(e.config.ElasticIndexes.Orders).Id(orderID).FetchSource(true).Do(ctx)
	if err != nil {
		tracing.TraceErr(span, err)
		return nil, errors.Wrap(err, "elasticClient.Get")
	}

	jsonData, err := result.Source.MarshalJSON()
	if err != nil {
		tracing.TraceErr(span, err)
		return nil, errors.Wrap(err, "Source.MarshalJSON")
	}

	var order models.OrderProjection
	if err := json.Unmarshal(jsonData, &order); err != nil {
		tracing.TraceErr(span, err)
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return &order, nil
}

func (e ElasticRepository) UpdateOrder(ctx context.Context, order *models.OrderProjection) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepository.UpdateShoppingCart")
	defer span.Finish()
	span.LogFields(log.String("OrderID", order.OrderID))

	_, err := e.elasticClient.Update().Index(e.config.ElasticIndexes.Orders).Id(order.OrderID).Doc(order).FetchSource(false).Do(ctx)
	if err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "elasticClient.Update")
	}

	return nil
}

func (e ElasticRepository) Search(ctx context.Context, text string, pq *utils.Pagination) (*dto.OrderSearchResponseDto, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "elasticRepository.Search")
	defer span.Finish()
	span.LogFields(log.String("Search", text))

	shouldMatch := v7.NewBoolQuery().
		Should(v7.NewMatchPhrasePrefixQuery(shopItemTitle, text), v7.NewMatchPhrasePrefixQuery(shopItemDescription, text)).
		MinimumNumberShouldMatch(minimumNumberShouldMatch)

	searchResult, err := e.elasticClient.Search(e.config.ElasticIndexes.Orders).
		Query(shouldMatch).
		From(pq.GetOffset()).
		Explain(e.config.Elastic.Explain).
		FetchSource(e.config.Elastic.FetchSource).
		Version(e.config.Elastic.Version).
		Size(pq.GetSize()).
		Pretty(e.config.Elastic.Pretty).
		Do(ctx)
	if err != nil {
		tracing.TraceErr(span, err)
		return nil, errors.Wrap(err, "elasticClient.Search")
	}

	orders := make([]*models.OrderProjection, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		jsonBytes, err := hit.Source.MarshalJSON()
		if err != nil {
			tracing.TraceErr(span, err)
			return nil, errors.Wrap(err, "Source.MarshalJSON")
		}
		var order models.OrderProjection
		if err := json.Unmarshal(jsonBytes, &order); err != nil {
			tracing.TraceErr(span, err)
			return nil, errors.Wrap(err, "json.Unmarshal")
		}
		orders = append(orders, &order)
	}

	return &dto.OrderSearchResponseDto{
		Pagination: dto.Pagination{
			TotalCount: searchResult.TotalHits(),
			TotalPages: int64(pq.GetTotalPages(int(searchResult.TotalHits()))),
			Page:       int64(pq.GetPage()),
			Size:       int64(pq.GetSize()),
			HasMore:    pq.GetHasMore(int(searchResult.TotalHits())),
		},
		Orders: utils.OrdersResponseFrom(orders),
	}, nil
}
