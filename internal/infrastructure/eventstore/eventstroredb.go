package eventstore

import (
	"github.com/EventStore/EventStore-Client-Go/esdb"
)

func NewEventStoreClient(cfg EventStoreConfig) (*esdb.Client, error) {
	settings, err := esdb.ParseConnectionString(cfg.ConnectionString)
	if err != nil {
		return nil, err
	}

	return esdb.NewClient(settings)
}
