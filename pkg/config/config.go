package config

import (
	"flag"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/wassef911/eventually/internal/infrastructure/elasticsearch"
	"github.com/wassef911/eventually/internal/infrastructure/eventstore"
	"github.com/wassef911/eventually/internal/infrastructure/mongodb"
	"github.com/wassef911/eventually/internal/infrastructure/tracing"
	"github.com/wassef911/eventually/pkg/logger"
)

type Config struct {
	ServiceName      string                      `mapstructure:"serviceName"`
	Logger           *logger.Config              `mapstructure:"logger"`
	Mongo            *mongodb.Config             `mapstructure:"mongo"`
	MongoCollections MongoCollections            `mapstructure:"mongoCollections"`
	Jaeger           *tracing.Config             `mapstructure:"jaeger"`
	EventStoreConfig eventstore.EventStoreConfig `mapstructure:"eventStoreConfig"`
	Subscriptions    Subscriptions               `mapstructure:"subscriptions"`
	Elastic          elasticsearch.Config        `mapstructure:"elastic"`
	ElasticIndexes   ElasticIndexes              `mapstructure:"elasticIndexes"`
	Port             string                      `mapstructure:"port" validate:"required"`
	Development      bool                        `mapstructure:"development"`
	BasePath         string                      `mapstructure:"basePath" validate:"required"`
}

type MongoCollections struct {
	Orders string `mapstructure:"orders" validate:"required"`
}

type Subscriptions struct {
	PoolSize                   int    `mapstructure:"poolSize" validate:"required,gte=0"`
	OrderPrefix                string `mapstructure:"orderPrefix" validate:"required,gte=0"`
	MongoProjectionGroupName   string `mapstructure:"mongoProjectionGroupName" validate:"required,gte=0"`
	ElasticProjectionGroupName string `mapstructure:"elasticProjectionGroupName" validate:"required,gte=0"`
}

type ElasticIndexes struct {
	Orders string `mapstructure:"orders" validate:"required"`
}

func New() (*Config, error) {
	var configPath string
	flag.StringVar(&configPath, "config", "cmd/server/config.yaml", "path to config file")

	cfg := &Config{}
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "viper.ReadInConfig")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, errors.Wrap(err, "viper.Unmarshal")
	}
	return cfg, nil
}
