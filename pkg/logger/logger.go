package logger

import (
	"os"
	"time"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/wassef911/eventually/internal/api/constants"
)

type Config struct {
	LogLevel string `mapstructure:"level"`
	Debug    bool   `mapstructure:"Debug"`
	Encoder  string `mapstructure:"encoder"`
}

// Logger methods interface
type Logger interface {
	InitLogger()
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	WarnMsg(msg string, err error)
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Err(msg string, err error)
	Fatal(args ...interface{})
	Printf(template string, args ...interface{})
	WithName(name string)
	HttpMiddlewareAccessLogger(method string, uri string, status int, size int64, time time.Duration)
	ProjectionEvent(projectionName string, groupName string, event *esdb.ResolvedEvent, workerID int)
}

// Application logger
type appLogger struct {
	level       string
	Debug       bool
	encoding    string
	sugarLogger *zap.SugaredLogger
	logger      *zap.Logger
}

// NewAppLogger App Logger constructor
func NewAppLogger(config *Config) *appLogger {
	return &appLogger{level: config.LogLevel, Debug: config.Debug, encoding: config.Encoder}
}

var loggerLevelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

func (l *appLogger) getLoggerLevel() zapcore.Level {
	level, exist := loggerLevelMap[l.level]
	if !exist {
		return zapcore.DebugLevel
	}

	return level
}

// InitLogger Init logger
func (l *appLogger) InitLogger() {
	logLevel := l.getLoggerLevel()

	logWriter := zapcore.AddSync(os.Stdout)

	var encoderCfg zapcore.EncoderConfig
	if l.Debug {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	var encoder zapcore.Encoder
	encoderCfg.NameKey = "[SERVICE]"
	encoderCfg.TimeKey = "[TIME]"
	encoderCfg.LevelKey = "[LEVEL]"
	encoderCfg.CallerKey = "[LINE]"
	encoderCfg.MessageKey = "[MESSAGE]"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	encoderCfg.EncodeDuration = zapcore.StringDurationEncoder

	if l.encoding == "console" {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderCfg.EncodeCaller = zapcore.FullCallerEncoder
		encoderCfg.ConsoleSeparator = " | "
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoderCfg.FunctionKey = "[CALLER]"
		encoderCfg.EncodeName = zapcore.FullNameEncoder
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(logLevel))
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	l.logger = logger
	l.sugarLogger = logger.Sugar()
}

// Logger methods

// WithName add logger microservice name
func (l *appLogger) WithName(name string) {
	l.logger = l.logger.Named(name)
	l.sugarLogger = l.sugarLogger.Named(name)
}

// Debugf uses fmt.Sprintf to log a templated message
func (l *appLogger) Debugf(template string, args ...interface{}) {
	l.sugarLogger.Debugf(template, args...)
}

// Info uses fmt.Sprint to construct and log a message
func (l *appLogger) Info(args ...interface{}) {
	l.sugarLogger.Info(args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (l *appLogger) Infof(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

// Printf uses fmt.Sprintf to log a templated message
func (l *appLogger) Printf(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func (l *appLogger) Warn(args ...interface{}) {
	l.sugarLogger.Warn(args...)
}

// WarnMsg log error message with warn level.
func (l *appLogger) WarnMsg(msg string, err error) {
	l.logger.Warn(msg, zap.String("error", err.Error()))
}

// Warnf uses fmt.Sprintf to log a templated message.
func (l *appLogger) Warnf(template string, args ...interface{}) {
	l.sugarLogger.Warnf(template, args...)
}

// Error uses fmt.Sprint to construct and log a message.
func (l *appLogger) Error(args ...interface{}) {
	l.sugarLogger.Error(args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (l *appLogger) Errorf(template string, args ...interface{}) {
	l.sugarLogger.Errorf(template, args...)
}

// Err uses error to log a message.
func (l *appLogger) Err(msg string, err error) {
	l.logger.Error(msg, zap.Error(err))
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (l *appLogger) Fatal(args ...interface{}) {
	l.sugarLogger.Fatal(args...)
}

func (l *appLogger) HttpMiddlewareAccessLogger(method, uri string, status int, size int64, time time.Duration) {
	l.logger.Info(
		constants.HTTP,
		zap.String(constants.METHOD, method),
		zap.String(constants.URI, uri),
		zap.Int(constants.STATUS, status),
		zap.Int64(constants.SIZE, size),
		zap.Duration(constants.TIME, time),
	)
}

func (l *appLogger) ProjectionEvent(projectionName string, groupName string, event *esdb.ResolvedEvent, workerID int) {
	l.logger.Info(
		projectionName,
		zap.String(constants.GroupName, groupName),
		zap.String(constants.StreamID, event.OriginalEvent().StreamID),
		zap.String(constants.EventID, event.OriginalEvent().EventID.String()),
		zap.String(constants.EventType, event.OriginalEvent().EventType),
		zap.Uint64(constants.EventNumber, event.OriginalEvent().EventNumber),
		zap.Time(constants.CreatedDate, event.OriginalEvent().CreatedDate),
		zap.String(constants.UserMetadata, string(event.OriginalEvent().UserMetadata)),
		zap.Int(constants.WorkerID, workerID),
	)
}
