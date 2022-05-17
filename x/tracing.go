package x

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

const tracingComponent = "github.com/ory/kratos"

// AddAttributes adds useful attributes to a given span.
func AddAttributes(r *http.Request, span trace.Span) {
	attrs := append(
		semconv.NetAttributesFromHTTPRequest("tcp", r),
		semconv.EndUserAttributesFromHTTPRequest(r)...,
	)
	attrs = append(attrs,
		semconv.HTTPServerAttributesFromHTTPRequest(r.URL.Path, "", r)...,
	)
	span.SetAttributes(attrs...)
}

// TraceHandle wraps httprouter.Handle with tracing
func TraceHandle(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		tracer := otel.GetTracerProvider().Tracer(tracingComponent)
		ctx, span := tracer.Start(r.Context(), r.URL.Path)
		AddAttributes(r, span)
		defer span.End()

		ctx = context.WithValue(ctx, "params", ps)
		r = r.WithContext(ctx)
		h(w, r, ps)
	}
}

// TraceHandlerFunc wraps http.HandlerFunc with tracing
func TraceHandlerFunc(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.GetTracerProvider().Tracer(tracingComponent)
		ctx, span := tracer.Start(r.Context(), r.URL.Path)
		AddAttributes(r, span)

		r = r.WithContext(ctx)
		h(w, r)
	}
}

// TraceHandler wraps http.Handler with tracing
func TraceHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.GetTracerProvider().Tracer(tracingComponent)
		ctx, span := tracer.Start(r.Context(), r.URL.Path)
		AddAttributes(r, span)
		defer span.End()

		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}
