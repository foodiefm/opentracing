package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
)

func TestWrapClient(t *testing.T) {
	tests := []struct {
		name string
		root bool
	}{
		{"no root span", false},
		{"root span", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// setup global tracer
			ct := mocktracer.New()
			opentracing.SetGlobalTracer(ct)
			defer opentracing.SetGlobalTracer(opentracing.NoopTracer{})

			// init tests server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.root {
					st := mocktracer.New()
					_, err := st.Extract(
						opentracing.HTTPHeaders,
						opentracing.HTTPHeadersCarrier(r.Header))
					if err != nil && err == opentracing.ErrSpanContextNotFound {
						t.Errorf("There is no span in request")
					}

				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}))
			defer server.Close()

			client := WrapClient(server.Client(), "")

			// create request
			ctx := context.Background()
			var rspan opentracing.Span
			if test.root {
				rspan = ct.StartSpan("root_span")
				ctx = opentracing.ContextWithSpan(ctx, rspan)
			}
			req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
			client.Do(req.WithContext(ctx))
			if test.root {
				rspan.Finish()
			}

			if !test.root {
				if len(ct.FinishedSpans()) > 0 {
					t.Error("There are spans even, there should not be")
				}
			} else {
				if len(ct.FinishedSpans()) != 2 {
					t.Error("There should be only root span and span created for request")
				}
			}
		})
	}
}
func TestExternalWrapClient(t *testing.T) {
	tests := []struct {
		name      string
		root      bool
		operation string
		connerr   bool
	}{
		{"no root span", false, "", false},
		{"no root span with connection error", false, "", true},
		{"no root span and opearation name given", false, "test.request", false},
		{"root span", true, "", false},
		{"root span with connection error", true, "", true},
		{"root span and opearation name given", true, "test.request", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// setup global tracer
			ct := mocktracer.New()
			opentracing.SetGlobalTracer(ct)
			defer opentracing.SetGlobalTracer(opentracing.NoopTracer{})

			// init tests server
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.root {
					st := mocktracer.New()
					_, err := st.Extract(
						opentracing.HTTPHeaders,
						opentracing.HTTPHeadersCarrier(r.Header))
					if err == nil || (err != nil && err != opentracing.ErrSpanContextNotFound) {
						t.Errorf("There is span in request or some unexpected error happen")
					}

				}
				if test.connerr {
					server.CloseClientConnections()
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}))
			defer server.Close()

			client := WrapExternalClient(server.Client(), test.operation)

			// create request
			ctx := context.Background()
			var rspan opentracing.Span
			if test.root {
				rspan = ct.StartSpan("root_span")
				ctx = opentracing.ContextWithSpan(ctx, rspan)
			}
			req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
			client.Do(req.WithContext(ctx))
			if test.root {
				rspan.Finish()
			}

			if !test.root {
				if len(ct.FinishedSpans()) > 0 {
					t.Error("There are spans even, there should not be")
				}
			} else {
				if len(ct.FinishedSpans()) != 2 {
					t.Error("There should be root span and request span")
				} else {
					span := ct.FinishedSpans()[0]

					expected := "http.request"
					if test.operation != "" {
						expected = test.operation
					}
					if expected != span.OperationName {
						t.Errorf("incorrect operation name")
					}
				}
			}
		})
	}
}

func ExampleWrapClient() {
	client := WrapClient(&http.Client{}, "span.name")

	span := opentracing.StartSpan("test_call")
	defer span.Finish()

	req, _ := http.NewRequest(http.MethodGet, "http://localhost/test", nil)
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))

	client.Do(req)
}

func ExampleWrapExternalClient() {
	client := WrapExternalClient(&http.Client{}, "external_call")

	span := opentracing.StartSpan("test_call")
	defer span.Finish()

	req, _ := http.NewRequest(http.MethodGet, "http://localhost/test", nil)
	ctx := opentracing.ContextWithSpan(context.TODO(), span)

	client.Do(req.WithContext(ctx))
}
