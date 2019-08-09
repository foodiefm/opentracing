package opentracer

import (
	"github.com/opentracing/opentracing-go"
	dd "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type opentracer struct {
	tracer      opentracing.Tracer
	serviceName string
}

// New creates, DataDog tracer instance and wraps it so
// that standard operation name and component definitions
// can be used
func New(opts ...ddtracer.StartOption) opentracing.Tracer {
	// FIXME: set service name
	return &opentracer{tracer: dd.New(opts...)}
}

// StartSpan wraps DataDog opentracing tracer StartSpan to change
// operation name and component name to be correct in Datadog perspective
func (o *opentracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	ddopts := []opentracing.StartSpanOption{}
	// Component name is now hardcoded to "http.request". This should be
	// changed to be dynamic later
	component := "http.request"

	ddopts = append(ddopts, opts...)
	ddopts = append(ddopts, dd.ResourceName(operationName))
	ddopts = append(ddopts, dd.ServiceName(o.serviceName))

	return o.tracer.StartSpan(component, ddopts...)
}

// Inject directly calls DataDog opentracer tracer Inject function
func (o *opentracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	return o.tracer.Inject(sm, format, carrier)
}

// Extract directly calls DataDog opentracer tracer Extract function
func (o *opentracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	span, err := o.tracer.Extract(format, carrier)
	if err == ddtracer.ErrSpanContextNotFound {
		err = opentracing.ErrSpanContextNotFound
	}

	return span, err
}
