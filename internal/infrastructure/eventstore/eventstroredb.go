package eventstore

import (
	"github.com/EventStore/EventStore-Client-Go/esdb"
)

func NewEventStoreClient(config EventStoreConfig) (*esdb.Client, error) {
	settings, err := esdb.ParseConnectionString(config.ConnectionString)
	if err != nil {
		return nil, err
	}

	return esdb.NewClient(settings)
}
