package gin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
)

func TestRequestTracer(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		externalSpan  string
		newTracer     func() *mocktracer.MockTracer
		injector      Inject
		operationName string
		status        int
		err           error
	}{
		{"No root span", "/test", "", mocktracer.New, nil, "GET__test", http.StatusOK, nil},
		{"No root span, error in handler", "/test", "", mocktracer.New, nil, "GET__test", http.StatusInternalServerError, fmt.Errorf("fatal error")},
		{"No root span and path params", "/test/123", "", mocktracer.New, nil, "GET__test_id", http.StatusOK, nil},
		{"Root span in request", "/test", "external", mocktracer.New, nil, "GET__test", http.StatusOK, nil},
		{"No root span and noop tracer", "/test", "", nil, nil, "GET__test", http.StatusOK, nil},
		{"Root span in request and noop tracer", "/test", "external", nil, nil, "GET__test", http.StatusOK, nil},
		{"Inject special tags", "/test", "", mocktracer.New, func(ctx context.Context, span opentracing.Span) context.Context {
			span.SetTag("test", 1)
			return ctx
		}, "GET__test", http.StatusOK, nil},
	}

	gin.SetMode(gin.ReleaseMode)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// init global tracer for middleware
			var externalTracer, mockTracer *mocktracer.MockTracer
			var span opentracing.Span
			opentracing.SetGlobalTracer(opentracing.NoopTracer{})
			if test.newTracer != nil {
				mockTracer = test.newTracer()
				opentracing.SetGlobalTracer(mockTracer)
				defer opentracing.SetGlobalTracer(opentracing.NoopTracer{})
			}

			// Create echo router and attach route
			router := gin.New()
			router.Use(RequestTracer(test.injector))
			handler := func(c *gin.Context) {
				if test.err != nil {
					c.AbortWithError(http.StatusInternalServerError, test.err)
				}

				c.String(http.StatusOK, "OK")
			}
			router.GET("/test", handler)
			router.GET("/test/:id", handler)

			// Create request and handle it
			req := httptest.NewRequest(http.MethodGet, test.path, nil)

			if test.externalSpan != "" {
				externalTracer = mocktracer.New()
				span = externalTracer.StartSpan(test.externalSpan)
				externalTracer.Inject(
					span.Context(),
					opentracing.HTTPHeaders,
					opentracing.HTTPHeadersCarrier(req.Header))
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			// assert results
			if rec.Code != test.status {
				t.Errorf("Not expected status code: %d", rec.Code)
			}

			if test.newTracer != nil {
				spans := mockTracer.FinishedSpans()
				if len(spans) != 1 {
					t.Errorf("incorrect number of spans")
					return
				}

				if spans[0].OperationName != test.operationName {
					t.Errorf("incorrect operation name")
				}

				if test.injector != nil {
					tags := spans[0].Tags()
					if v, exists := tags["test"]; !exists || v != 1 {
						t.Errorf("special tag not added")
					}
				}

				if test.err != nil {
					tags := spans[0].Tags()
					if v, exists := tags["server.errors"]; !exists || v != test.err.Error() {
						t.Errorf("special tag not added")
					}
				}

				if test.externalSpan != "" {
					span.Finish()
					if spans[0].ParentID != externalTracer.FinishedSpans()[0].SpanContext.SpanID {
						t.Error("Middleware span is not child of its root")
					}

					if spans[0].SpanContext.TraceID != externalTracer.FinishedSpans()[0].SpanContext.TraceID {
						t.Error("Middleware span does not use same trace id, than its parent")
					}
				}
			}
		})
	}
}
