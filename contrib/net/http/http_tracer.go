package http

import (
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// rounddTripper is extension to standard RoundTripper for internal calls.
type roundTripper struct {
	operationName string
	base          http.RoundTripper
}

// RoundTrip for internal calls does not start new span as it assumes that called
// service creates its own span to query.
func (rt *roundTripper) RoundTrip(req *http.Request) (res *http.Response, err error) {
	rootSpan := opentracing.SpanFromContext(req.Context())
	if rootSpan != nil {
		// context contains span, create new child span
		span, _ := opentracing.StartSpanFromContext(req.Context(), rt.operationName)
		ext.HTTPMethod.Set(span, req.Method)
		ext.HTTPUrl.Set(span, req.URL.Path)
		defer func() {
			if err == nil {
				ext.HTTPStatusCode.Set(span, uint16(res.StatusCode))
			} else {
				span.SetTag("server.errors", err.Error())
				ext.Error.Set(span, true)
			}
			span.Finish()
		}()

		opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header))
	}

	return rt.base.RoundTrip(req)
}

// WrapClient wraps http.Client to inject opentracing span to outgoing call
// if and only if root span exists in request context.
func WrapClient(c *http.Client, on string) *http.Client {
	rt := http.DefaultTransport
	if c.Transport != nil {
		rt = c.Transport
	}
	if on == "" {
		on = "http.request"
	}

	return &http.Client{
		Transport: &roundTripper{
			base:          rt,
			operationName: on,
		},
		CheckRedirect: c.CheckRedirect,
		Jar:           c.Jar,
	}
}

// externalRoundTripper in extension to standard RoundTripper and will create new span
// which is not injected to call to show those in our monitoring
type externalRoundTripper struct {
	base          http.RoundTripper
	operationName string
}

// RoundTrip creates new span, but don't inject it to request headers as this is intended to use
// with external services.
func (rt *externalRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var res *http.Response
	var err error

	ctx := req.Context()
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		// context contains span, create new child span
		span, ctx = opentracing.StartSpanFromContext(ctx, rt.operationName)
		ext.HTTPMethod.Set(span, req.Method)
		ext.HTTPUrl.Set(span, req.URL.Path)
		defer func() {
			if err == nil {
				ext.HTTPStatusCode.Set(span, uint16(res.StatusCode))
			} else {
				span.SetTag("server.errors", err.Error())
				ext.Error.Set(span, true)
			}
			span.Finish()
		}()
	}

	res, err = rt.base.RoundTrip(req.WithContext(ctx))

	return res, err
}

// WrapExternalClient wraps http.Client with tracing and creates
// opentracing span to outgoing calls, if and only if context
// of request contains root span. default operation name for that
// span is http.request
func WrapExternalClient(c *http.Client, on string) *http.Client {
	rt := http.DefaultTransport
	if c.Transport != nil {
		rt = c.Transport
	}
	if on == "" {
		on = "http.request"
	}

	return &http.Client{
		Transport: &externalRoundTripper{
			base:          rt,
			operationName: on,
		},
		CheckRedirect: c.CheckRedirect,
		Jar:           c.Jar,
	}
}
