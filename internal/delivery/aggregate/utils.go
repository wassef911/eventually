package aggregate

import (
	"context"
	"strings"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/wassef911/eventually/internal/delivery/models"
	"github.com/wassef911/eventually/internal/infrastructure/es"
	"github.com/wassef911/eventually/internal/infrastructure/es/store"
)

func GetShopItemsTotalPrice(shopItems []*models.ShopItem) float64 {
	var totalPrice float64 = 0
	for _, item := range shopItems {
		totalPrice += item.Price * float64(item.Quantity)
	}
	return totalPrice
}

// GetOrderAggregateID get order aggregate id for eventstoredb
func GetOrderAggregateID(eventAggregateID string) string {
	return strings.ReplaceAll(eventAggregateID, "order-", "")
}

func IsAggregateNotFound(aggregate es.Aggregate) bool {
	return aggregate.GetVersion() == 0
}

func LoadOrderAggregate(ctx context.Context, eventStore store.AggregateStore, aggregateID string) (*OrderAggregate, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "LoadOrderAggregate")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", aggregateID))

	order := NewOrderAggregateWithID(aggregateID)

	err := eventStore.Exists(ctx, order.GetID())
	if err != nil && !errors.Is(err, esdb.ErrStreamNotFound) {
		return nil, err
	}

	if err := eventStore.Load(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}
