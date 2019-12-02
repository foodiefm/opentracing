package pg

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
)

func TestTracingQueryHook(t *testing.T) {
	tests := []struct {
		name string
		err  bool
		root bool
	}{
		{"tracing when there is root span", false, true},
		{"do not trace when there is no root span", false, false},
		{"tracing when there is root span with error", true, true},
		{"do not trace when there is no root span with error", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracer := mocktracer.New()
			opentracing.SetGlobalTracer(tracer)
			ctx := context.Background()
			root := tracer.StartSpan("root")
			if tt.root {
				ctx = opentracing.ContextWithSpan(ctx, root)
			}

			event := &pg.QueryEvent{
				Query:     "SELECT 1 FROM tests",
				StartTime: time.Now(),
			}
			hook := TracingQueryHook{}
			ctx, err := hook.BeforeQuery(ctx, event)
			if err != nil {
				t.Errorf("TracingQueryHook.BeforeQuery() error = %v", err)
				return
			}
			if tt.err {
				event.Err = fmt.Errorf("operation failed")
			}
			err = hook.AfterQuery(ctx, event)
			root.Finish()

			spans := tracer.FinishedSpans()
			if tt.root {
				if len(spans) != 2 {
					t.Errorf("Incorrect number of spans, expected 2, got %d", len(spans))
				}
				span := spans[0]
				if span.OperationName != event.Query {
					t.Errorf("Incorrect operation name, expected: '%s', got: '%s'", event.Query, span.OperationName)
				}
				t.Log(span)
			} else {
				if len(spans) != 1 {
					t.Errorf("Incorrect number of spans, expected 1, got %d", len(spans))
				}
			}
		})
	}
}
