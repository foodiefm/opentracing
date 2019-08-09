package gin

import (
	"context"

	"github.com/foodiefm/opentracing/utils"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Inject typed functions can be used to inject
// implementation specific changes to span
type Inject func(context.Context, *opentracing.Span) context.Context

// RequestTracer implements gin middleware to trace requests
// using opentracing
func RequestTracer(inject Inject) gin.HandlerFunc {
	return func(c *gin.Context) {
		if opentracing.IsGlobalTracerRegistered() {
			req := c.Request

			wireContext, err := opentracing.GlobalTracer().Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(req.Header))
			if err != nil && err == opentracing.ErrSpanContextNotFound {
				// Log potential tracing errors
			}

			// list path parameters and parameter names
			params := map[string]string{}
			for _, p := range c.Params {
				params[p.Value] = p.Key
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

				ext.HTTPStatusCode.Set(serverSpan, uint16(c.Writer.Status()))
				if len(c.Errors) > 0 {
					serverSpan.SetTag("server.errors", c.Errors.String())
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
				ctx = inject(ctx, &serverSpan)
			}

			req = req.WithContext(ctx)

			// Add updated context to request
			c.Request = req
		}
		c.Next()
	}
}
