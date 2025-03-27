package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	v7 "github.com/olivere/elastic/v7"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/wassef911/eventually/internal/api/constants"
	"github.com/wassef911/eventually/internal/api/middlewares"
	"github.com/wassef911/eventually/internal/delivery/projections/elastic"
	"github.com/wassef911/eventually/internal/delivery/projections/mongo"
	"github.com/wassef911/eventually/internal/delivery/repository"
	service "github.com/wassef911/eventually/internal/delivery/services"
	"github.com/wassef911/eventually/internal/infrastructure/elasticsearch"
	"github.com/wassef911/eventually/internal/infrastructure/es/store"
	"github.com/wassef911/eventually/internal/infrastructure/eventstore"
	"github.com/wassef911/eventually/internal/infrastructure/mongodb"
	"github.com/wassef911/eventually/internal/infrastructure/tracing"
	"github.com/wassef911/eventually/pkg/config"
	"github.com/wassef911/eventually/pkg/logger"
)

const (
	maxHeaderBytes       = 1 << 20
	stackSize            = 1 << 10 // 1 KB
	bodyLimit            = "2M"
	readTimeout          = 15 * time.Second
	writeTimeout         = 15 * time.Second
	gzipLevel            = 5
	waitShotDownDuration = 3 * time.Second
)

type IServer interface {
	Run() error
	configureRoutes()
	initDatabase(ctx context.Context)
	initEngine(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

var _ IServer = &server{}

type server struct {
	cfg           *config.Config
	log           logger.Logger
	mw            middlewares.MiddlewareManager
	os            *service.OrderService
	v             *validator.Validate
	mongoClient   *mongoDriver.Client
	elasticClient *v7.Client
	echo          *echo.Echo
	ps            *http.Server
	doneCh        chan struct{}
}

func New(cfg *config.Config, log logger.Logger) *server {
	return &server{cfg: cfg, log: log, v: validator.New(), echo: echo.New(), mw: middlewares.NewMiddlewareManager(log, cfg), doneCh: make(chan struct{})}
}

func (s *server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err := s.v.StructCtx(ctx, s.cfg); err != nil {
		return errors.Wrap(err, "cfg validate")
	}

	tracer, closer, err := tracing.New(s.cfg.Jaeger)
	if err != nil {
		return err
	}
	defer closer.Close() // nolint: errcheck
	opentracing.SetGlobalTracer(tracer)

	mongoDBConn, err := mongodb.NewMongoClient(ctx, s.cfg.Mongo)
	if err != nil {
		return errors.Wrap(err, "mongodb connection error"+s.cfg.Mongo.URI)
	}
	s.mongoClient = mongoDBConn
	defer mongoDBConn.Disconnect(ctx) // nolint: errcheck
	s.log.Infof("(Mongo connected) SessionsInProgress: {%v}", mongoDBConn.NumberSessionsInProgress())

	if err := s.initEngine(ctx); err != nil {
		s.log.Errorf("(initEngine) err: {%v}", err)
		return err
	}

	mongoRepository := repository.NewMongoRepository(s.log, s.cfg, s.mongoClient)
	elasticRepository := repository.NewElasticRepository(s.log, s.cfg, s.elasticClient)

	db, err := eventstore.NewEventStoreClient(s.cfg.EventStoreConfig)
	if err != nil {
		return err
	}
	defer db.Close() // nolint: errcheck

	aggregateStore := store.NewAggregateStore(s.log, db)
	s.os = service.New(s.log, s.cfg, aggregateStore, mongoRepository, elasticRepository)

	mongoProjection := mongo.NewOrderProjection(s.log, db, mongoRepository, s.cfg)
	elasticProjection := elastic.NewElasticProjection(s.log, db, elasticRepository, s.cfg)

	go func() {
		err := mongoProjection.Subscribe(ctx, []string{s.cfg.Subscriptions.OrderPrefix}, s.cfg.Subscriptions.PoolSize, mongoProjection.ProcessEvents)
		if err != nil {
			s.log.Errorf("(orderProjection.Subscribe) err: {%v}", err)
			cancel()
		}
	}()

	go func() {
		err := elasticProjection.Subscribe(ctx, []string{s.cfg.Subscriptions.OrderPrefix}, s.cfg.Subscriptions.PoolSize, elasticProjection.ProcessEvents)
		if err != nil {
			s.log.Errorf("(elasticProjection.Subscribe) err: {%v}", err)
			cancel()
		}
	}()

	s.initDatabase(ctx)

	go func() {
		s.configureRoutes()
		s.echo.Server.ReadTimeout = readTimeout
		s.echo.Server.WriteTimeout = writeTimeout
		s.echo.Server.MaxHeaderBytes = maxHeaderBytes
		if err := s.echo.Start(s.cfg.Port); err != nil {
			s.log.Errorf("Error Starting the server: {%v}", err)
			cancel()
		}
	}()
	s.log.Infof("%s is listening on PORT: {%s}", s.cfg.ServiceName, s.cfg.Port)

	<-ctx.Done()
	if err := s.Shutdown(ctx); err != nil {
		s.log.Warnf("(shutDownHealthCheckServer) err: {%v}", err)
	}

	<-s.doneCh
	s.log.Infof("%s server exited properly", s.cfg.ServiceName)
	return nil
}

func (s *server) initDatabase(ctx context.Context) {
	err := s.mongoClient.Database(s.cfg.Mongo.Db).CreateCollection(ctx, s.cfg.MongoCollections.Orders)
	if err != nil {
		s.log.Warnf("(CreateCollection) err: {%v}", err)
	}

	indexOptions := options.Index().SetSparse(true).SetUnique(true)
	index, err := s.mongoClient.Database(s.cfg.Mongo.Db).Collection(s.cfg.MongoCollections.Orders).Indexes().CreateOne(ctx, mongoDriver.IndexModel{
		Keys:    bson.D{{Key: constants.OrderIdIndex, Value: 1}},
		Options: indexOptions,
	})
	if err != nil {
		s.log.Warnf("(CreateOne) err: {%v}", err)
	}
	s.log.Infof("(CreatedIndex) index: {%s}", index)

	list, err := s.mongoClient.Database(s.cfg.Mongo.Db).Collection(s.cfg.MongoCollections.Orders).Indexes().List(ctx)
	if err != nil {
		s.log.Warnf("(initDatabase) [List] err: {%v}", err)
	}

	if list != nil {
		var results []bson.M
		if err := list.All(ctx, &results); err != nil {
			s.log.Warnf("(All) err: {%v}", err)
		}
		s.log.Infof("(indexes) results: {%#v}", results)
	}

	collections, err := s.mongoClient.Database(s.cfg.Mongo.Db).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		s.log.Warnf("(ListCollections) err: {%v}", err)
	}
	s.log.Infof("(Collections) created collections: {%v}", collections)
}

func (s *server) initEngine(ctx context.Context) error {
	elasticClient, err := elasticsearch.NewElasticClient(s.cfg.Elastic)
	if err != nil {
		return err
	}
	s.elasticClient = elasticClient

	info, code, err := s.elasticClient.Ping(s.cfg.Elastic.URL).Do(ctx)
	if err != nil {
		return errors.Wrap(err, "client.Ping")
	}
	s.log.Infof("Elasticsearch returned with code {%d} and version {%s}", code, info.Version.Number)

	esVersion, err := s.elasticClient.ElasticsearchVersion(s.cfg.Elastic.URL)
	if err != nil {
		return errors.Wrap(err, "client.ElasticsearchVersion")
	}
	s.log.Infof("Elasticsearch version {%s}", esVersion)

	return nil
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.ps.Shutdown(ctx)
}
