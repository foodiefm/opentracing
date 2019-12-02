package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

// TracingQueryHook to add opentracing to pg package
type TracingQueryHook struct{}

// BeforeQuery hook stores start time of request
func (t TracingQueryHook) BeforeQuery(ctx context.Context, e *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

// AfterQuery hook creates span with unformatted query and start time stored by BeforeQuery hook
func (t TracingQueryHook) AfterQuery(ctx context.Context, e *pg.QueryEvent) error {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		q, err := e.UnformattedQuery()
		if err != nil {
			return nil
		}

		start := opentracing.StartTime(e.StartTime)
		sp := opentracing.StartSpan(
			q,
			opentracing.ChildOf(span.Context()),
			start,
		)

		ext.DBType.Set(sp, "sql")
		ext.DBStatement.Set(sp, q)
		if e.Err != nil {
			ext.Error.Set(sp, true)
			sp.LogFields(
				log.String("event", "error"),
				log.String("message", e.Err.Error()))
		}
		defer sp.Finish()
	}
	return nil
}
