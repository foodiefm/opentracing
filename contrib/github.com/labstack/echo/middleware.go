package echo

import (
	"context"

	"github.com/foodiefm/opentracing/utils"
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Inject typed functions can be used to inject
// implementation specific changes to span
type Inject func(context.Context, opentracing.Span) context.Context

// RequestTracer created middleware to trace requests using opentracing
func RequestTracer(inject Inject) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var err error
			if opentracing.IsGlobalTracerRegistered() {
				req := c.Request()

				wireContext, eerr := opentracing.GlobalTracer().Extract(
					opentracing.HTTPHeaders,
					opentracing.HTTPHeadersCarrier(req.Header))
				if eerr != nil && eerr == opentracing.ErrSpanContextNotFound {
					// Log potential tracing errors
				}

				// list path parameters and parameter names
				params := map[string]string{}
				for _, name := range c.ParamNames() {
					params[c.Param(name)] = name
				}

				// Create server span or create new root span
				serverSpan, ctx := opentracing.StartSpanFromContext(
					req.Context(),
					utils.GetResourceName(req, params),
					ext.RPCServerOption(wireContext),
				)

				// Ensure that span is finished and return status is added to it
				defer func() {
					defer serverSpan.Finish()

					ext.HTTPStatusCode.Set(serverSpan, uint16(c.Response().Status))
					if err != nil {
						serverSpan.SetTag("server.errors", err.Error())
						ext.Error.Set(serverSpan, true)
					}
				}()

				// Add tags to span
				ext.HTTPMethod.Set(serverSpan, req.Method)
				ext.HTTPUrl.Set(serverSpan, req.URL.Path)
				serverSpan.SetTag("span.type", "web")

				// Add span to Request object Context
				ctx = opentracing.ContextWithSpan(ctx, serverSpan)

				// Inject specific handling to spans
				if inject != nil {
					ctx = inject(ctx, serverSpan)
				}

				req = req.WithContext(ctx)

				// Add updated context to request
				c.SetRequest(req)
			}

			// Assign to err, so that value is usable
			// in deferred span finish call.
			err = next(c)
			return err
		}
	}
}
