package middlewares

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go/log"

	"github.com/wassef911/eventually/internal/infrastructure/tracing"
	"github.com/wassef911/eventually/pkg/config"
	"github.com/wassef911/eventually/pkg/errors"
	"github.com/wassef911/eventually/pkg/logger"
)

type MiddlewareManager interface {
	Apply(next echo.HandlerFunc) echo.HandlerFunc
}

type middlewareManager struct {
	log logger.Logger
	cfg *config.Config
}

func NewMiddlewareManager(log logger.Logger, cfg *config.Config) *middlewareManager {
	return &middlewareManager{log: log, cfg: cfg}
}

func (mw *middlewareManager) Apply(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		start := time.Now()
		res := ctx.Response()
		req := ctx.Request()
		operationName := fmt.Sprintf("%s %s", ctx.Request().Method, ctx.Path())
		newCtx, span := tracing.StartHttpServerTracerSpan(ctx, operationName)
		span.LogFields(
			log.String("RemoteAddr", req.RemoteAddr),
			log.String("Path", req.URL.Path),
		)

		newReq := req.WithContext(newCtx)
		ctx.SetRequest(newReq)

		err := next(ctx)
		span.Finish() // Finish AFTER downstream handlers complete

		status := res.Status
		size := res.Size
		s := time.Since(start)

		// log it
		mw.log.HttpMiddlewareAccessLogger(req.Method, newReq.URL.Path, status, size, s)

		// stop trace
		if err != nil {
			tracing.TraceErr(span, err)
			return errors.ErrorCtxResponse(ctx, err, mw.cfg.Logger.Debug)
		}
		return err
	}
}
