package middlewares

import (
	"time"

	"github.com/labstack/echo/v4"

	"github.com/wassef911/astore/internal/infrastructure/tracing"
	"github.com/wassef911/astore/pkg/config"
	"github.com/wassef911/astore/pkg/errors"
	"github.com/wassef911/astore/pkg/logger"
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
		req := ctx.Request()
		res := ctx.Response()
		cc, span := tracing.StartHttpServerTracerSpan(ctx, req.URL.String())
		defer span.Finish()
		// Store span in context for access in handler
		ctx.SetRequest(ctx.Request().WithContext(cc))

		err := next(ctx)

		status := res.Status
		size := res.Size
		s := time.Since(start)

		// log it
		mw.log.HttpMiddlewareAccessLogger(req.Method, req.URL.String(), status, size, s)

		// stop trace
		if err != nil {
			tracing.TraceErr(span, err)
			return errors.ErrorCtxResponse(ctx, err, mw.cfg.Logger.Debug)
		}
		return err
	}
}
