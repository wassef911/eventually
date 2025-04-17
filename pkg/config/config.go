package config

import (
	"strings"

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
	// Set up viper to read from environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind all environent variables
	bindEnvVars()

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, errors.Wrap(err, "viper.Unmarshal")
	}
	return config, nil
}
func bindEnvVars() {
	// Service Configuration
	viper.BindEnv("servicename", "SERVICE_NAME") // matches mapstructure:"servicename"
	viper.BindEnv("port", "PORT")
	viper.BindEnv("development", "DEVELOPMENT")
	viper.BindEnv("basepath", "BASE_PATH")

	// Logger Configuration
	viper.BindEnv("logger.level", "LOGGER_LEVEL")
	viper.BindEnv("logger.debug", "LOGGER_DEBUG")
	viper.BindEnv("logger.encoder", "LOGGER_ENCODER")

	// MongoDB Configuration
	viper.BindEnv("mongo.uri", "MONGO_URI")
	viper.BindEnv("mongo.user", "MONGO_USER")
	viper.BindEnv("mongo.password", "MONGO_PASSWORD")
	viper.BindEnv("mongo.db", "MONGO_DB")
	viper.BindEnv("mongocollections.orders", "MONGO_COLLECTIONS_ORDERS")

	// Jaeger Configuration
	viper.BindEnv("jaeger.enable", "JAEGER_ENABLE")
	viper.BindEnv("jaeger.servicename", "JAEGER_SERVICE_NAME")
	viper.BindEnv("jaeger.hostport", "JAEGER_HOST_PORT")
	viper.BindEnv("jaeger.logspans", "JAEGER_LOG_SPANS")

	// EventStore Configuration
	viper.BindEnv("eventstoreconfig.connectionstring", "EVENTSTORE_CONFIG_CONNECTION_STRING")

	// Subscriptions Configuration
	viper.BindEnv("subscriptions.poolsize", "SUBSCRIPTIONS_POOL_SIZE")
	viper.BindEnv("subscriptions.orderprefix", "SUBSCRIPTIONS_ORDER_PREFIX")
	viper.BindEnv("subscriptions.mongoprojectiongroupname", "SUBSCRIPTIONS_MONGO_PROJECTION_GROUP_NAME")
	viper.BindEnv("subscriptions.elasticprojectiongroupname", "SUBSCRIPTIONS_ELASTIC_PROJECTION_GROUP_NAME")

	// ElasticSearch Configuration
	viper.BindEnv("elastic.url", "ELASTIC_URL")
	viper.BindEnv("elastic.sniff", "ELASTIC_SNIFF")
	viper.BindEnv("elastic.gzip", "ELASTIC_GZIP")
	viper.BindEnv("elastic.explain", "ELASTIC_EXPLAIN")
	viper.BindEnv("elastic.fetchsource", "ELASTIC_FETCH_SOURCE")
	viper.BindEnv("elastic.version", "ELASTIC_VERSION")
	viper.BindEnv("elastic.pretty", "ELASTIC_PRETTY")
	viper.BindEnv("elasticindexes.orders", "ELASTIC_INDEXES_ORDERS")
}
