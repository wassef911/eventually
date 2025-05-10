package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	v7 "github.com/olivere/elastic/v7"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.mongodb.org/mongo-driver/bson"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/wassef911/eventually/docs"
	"github.com/wassef911/eventually/internal/api/constants"
	"github.com/wassef911/eventually/internal/api/handlers"
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

type Server struct {
	config        *config.Config
	log           logger.Logger
	mw            middlewares.MiddlewareManager
	orderService  *service.OrderService
	validator     *validator.Validate
	mongoClient   *mongoDriver.Client
	elasticClient *v7.Client
	echo          *echo.Echo
	httpServer    *http.Server
	doneCh        chan struct{}
}

func New(config *config.Config, log logger.Logger) *Server {
	return &Server{
		config:    config,
		log:       log,
		validator: validator.New(),
		echo:      echo.New(),
		mw:        middlewares.NewMiddlewareManager(log, config),
		doneCh:    make(chan struct{}),
	}
}

func (s *Server) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := s.validateConfig(ctx); err != nil {
		return err
	}

	tracer, closer, err := tracing.New(s.config.Jaeger)
	if err != nil {
		return err
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	if err := s.setupDatabases(ctx); err != nil {
		return err
	}

	mongoRepo := repository.NewMongoRepository(s.log, s.config, s.mongoClient)
	elasticRepo := repository.NewElasticRepository(s.log, s.config, s.elasticClient)

	db, err := eventstore.NewEventStoreClient(s.config.EventStoreConfig)
	if err != nil {
		return err
	}
	defer db.Close()

	aggregateStore := store.NewAggregateStore(s.log, db)
	s.orderService = service.New(s.log, s.config, aggregateStore, mongoRepo, elasticRepo)
	mongoProjection := mongo.NewOrderProjection(s.log, db, *mongoRepo, s.config)
	elasticProjection := elastic.NewElasticProjection(s.log, db, elasticRepo, s.config)
	go func() {
		err := mongoProjection.Subscribe(ctx, []string{s.config.Subscriptions.OrderPrefix}, s.config.Subscriptions.PoolSize, mongoProjection.ProcessEvents)
		if err != nil {
			s.log.Errorf("(orderProjection.Subscribe) err: {%v}", err)
			stop()
		}
	}()

	go func() {
		err := elasticProjection.Subscribe(ctx, []string{s.config.Subscriptions.OrderPrefix}, s.config.Subscriptions.PoolSize, elasticProjection.ProcessEvents)
		if err != nil {
			s.log.Errorf("(elasticProjection.Subscribe) err: {%v}", err)
			stop()
		}
	}()

	s.configureServer()
	s.log.Infof("%s is listening on PORT: {%s}", s.config.ServiceName, s.config.Port)
	if err := s.echo.Start(s.config.Port); err != nil {
		stop()
		return err
	}

	s.waitForShutdown(ctx)
	return nil
}

func (s *Server) validateConfig(ctx context.Context) error {
	if err := s.validator.StructCtx(ctx, s.config); err != nil {
		return errors.Wrap(err, "config validate")
	}
	return nil
}

func (s *Server) setupDatabases(ctx context.Context) error {
	if err := s.setupMongoDB(ctx); err != nil {
		return err
	}

	if err := s.initEngine(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Server) setupMongoDB(ctx context.Context) error {
	mongoDBConn, err := mongodb.NewMongoClient(ctx, s.config.Mongo)
	if err != nil {
		return errors.Wrap(err, "mongodb connection error"+s.config.Mongo.URI)
	}
	s.mongoClient = mongoDBConn
	defer mongoDBConn.Disconnect(ctx)

	s.initMongoCollections(ctx)
	return nil
}

func (s *Server) initEngine(ctx context.Context) error {
	elasticClient, err := elasticsearch.NewElasticClient(s.config.Elastic)
	if err != nil {
		return err
	}
	s.elasticClient = elasticClient

	info, code, err := s.elasticClient.Ping(s.config.Elastic.URL).Do(ctx)
	if err != nil {
		return errors.Wrap(err, "client.Ping")
	}
	s.log.Infof("Elasticsearch returned with code {%d} and version {%s}", code, info.Version.Number)

	esVersion, err := s.elasticClient.ElasticsearchVersion(s.config.Elastic.URL)
	if err != nil {
		return errors.Wrap(err, "client.ElasticsearchVersion")
	}
	s.log.Infof("Elasticsearch version {%s}", esVersion)

	return nil
}

func (s *Server) configureServer() {
	s.setupAPIHandlers()
	s.setupSwagger()
	s.setupGlobalMiddlewares()
	s.echo.Server.ReadTimeout = readTimeout
	s.echo.Server.WriteTimeout = writeTimeout
	s.echo.Server.MaxHeaderBytes = maxHeaderBytes
}

func (s *Server) waitForShutdown(ctx context.Context) {
	<-ctx.Done()
	if err := s.Shutdown(ctx); err != nil {
		s.log.Warnf("(shutDownHealthCheckServer) err: {%v}", err)
	}

	<-s.doneCh
	s.log.Infof("%s server exited properly", s.config.ServiceName)
}

func (s *Server) initMongoCollections(ctx context.Context) {
	err := s.mongoClient.Database(s.config.Mongo.Db).CreateCollection(ctx, s.config.MongoCollections.Orders)
	if err != nil {
		s.log.Warnf("(CreateCollection) err: {%v}", err)
	}

	indexOptions := options.Index().SetSparse(true).SetUnique(true)
	index, err := s.mongoClient.Database(s.config.Mongo.Db).Collection(s.config.MongoCollections.Orders).Indexes().CreateOne(ctx, mongoDriver.IndexModel{
		Keys:    bson.D{{Key: constants.OrderIdIndex, Value: 1}},
		Options: indexOptions,
	})
	if err != nil {
		s.log.Warnf("(CreateOne) err: {%v}", err)
	}
	s.log.Infof("(CreatedIndex) index: {%s}", index)

	list, err := s.mongoClient.Database(s.config.Mongo.Db).Collection(s.config.MongoCollections.Orders).Indexes().List(ctx)
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

	collections, err := s.mongoClient.Database(s.config.Mongo.Db).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		s.log.Warnf("(ListCollections) err: {%v}", err)
	}
	s.log.Infof("(Collections) created collections: {%v}", collections)
}

func (s *Server) setupAPIHandlers() {
	orderHandlers := handlers.NewOrderHandlers(
		s.echo.Group("/api/orders"),
		s.log,
		s.mw,
		s.config,
		s.validator,
		s.orderService,
	)
	orderHandlers.MapRoutes()
}

func (s *Server) setupSwagger() {
	docs.SwaggerInfo_swagger.Version = "1.0"
	docs.SwaggerInfo_swagger.BasePath = "/api"
	s.echo.GET("/swagger/*", echoSwagger.WrapHandler)
}

func (s *Server) setupGlobalMiddlewares() {
	s.echo.Use(
		middleware.RecoverWithConfig(middleware.RecoverConfig{
			StackSize:         stackSize,
			DisablePrintStack: true,
			DisableStackAll:   true,
		}),
		middleware.RequestID(),
		s.createGzipMiddleware(),
		middleware.BodyLimit(bodyLimit),
		s.mw.Apply,
	)
}

func (s *Server) createGzipMiddleware() echo.MiddlewareFunc {
	return middleware.GzipWithConfig(middleware.GzipConfig{
		Level: gzipLevel,
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Request().URL.Path, "swagger")
		},
	})
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
