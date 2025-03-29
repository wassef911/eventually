package tracing

import (
	"context"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/wassef911/eventually/internal/infrastructure/es"
)

func StartHttpServerTracerSpan(c echo.Context, operationName string) (context.Context, opentracing.Span) {
	req := c.Request()
	spanCtx, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	var serverSpan opentracing.Span
	if err != nil {
		// No parent span found in headers, start a new trace
		serverSpan = opentracing.GlobalTracer().StartSpan(operationName)
	} else {
		// Continue the existing trace
		serverSpan = opentracing.GlobalTracer().StartSpan(
			operationName,
			ext.RPCServerOption(spanCtx),
			opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
			opentracing.Tag{Key: string(ext.SpanKind), Value: ext.SpanKindRPCServerEnum},
		)
	}

	// Add standard HTTP tags
	ext.HTTPMethod.Set(serverSpan, req.Method)
	ext.HTTPUrl.Set(serverSpan, req.URL.Path)

	ctx := opentracing.ContextWithSpan(req.Context(), serverSpan)
	return ctx, serverSpan
}

func GetTextMapCarrierFromEvent(event es.Event) opentracing.TextMapCarrier {
	metadataMap := make(opentracing.TextMapCarrier)
	err := json.Unmarshal(event.GetMetadata(), &metadataMap)
	if err != nil {
		return metadataMap
	}
	return metadataMap
}

func StartProjectionTracerSpan(ctx context.Context, operationName string, event es.Event) (context.Context, opentracing.Span) {
	textMapCarrierFromMetaData := GetTextMapCarrierFromEvent(event)

	span, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, textMapCarrierFromMetaData)
	if err != nil {
		serverSpan := opentracing.GlobalTracer().StartSpan(operationName)
		ctx = opentracing.ContextWithSpan(ctx, serverSpan)
		return ctx, serverSpan
	}

	serverSpan := opentracing.GlobalTracer().StartSpan(operationName, ext.RPCServerOption(span))
	ctx = opentracing.ContextWithSpan(ctx, serverSpan)

	return ctx, serverSpan
}

func InjectTextMapCarrier(spanCtx opentracing.SpanContext) (opentracing.TextMapCarrier, error) {
	m := make(opentracing.TextMapCarrier)
	if err := opentracing.GlobalTracer().Inject(spanCtx, opentracing.TextMap, m); err != nil {
		return nil, err
	}
	return m, nil
}

func ExtractTextMapCarrier(spanCtx opentracing.SpanContext) opentracing.TextMapCarrier {
	textMapCarrier, err := InjectTextMapCarrier(spanCtx)
	if err != nil {
		return make(opentracing.TextMapCarrier)
	}
	return textMapCarrier
}

func ExtractTextMapCarrierBytes(spanCtx opentracing.SpanContext) []byte {
	textMapCarrier, err := InjectTextMapCarrier(spanCtx)
	if err != nil {
		return []byte("")
	}

	dataBytes, err := json.Marshal(&textMapCarrier)
	if err != nil {
		return []byte("")
	}
	return dataBytes
}

func TraceErr(span opentracing.Span, err error) {
	span.SetTag("error", true)
	span.LogKV("error_code", err.Error())
}
