package main

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	tp := sdktrace.NewTracerProvider()
	ctx, span := tp.Tracer("x").Start(context.Background(), "op")
	defer span.End()

	// o que viaja entre servicos
	h := http.Header{}
	propagation.TraceContext{}.Inject(ctx, propagation.HeaderCarrier(h))
	fmt.Println("traceparent:", h.Get("Traceparent"))
	fmt.Println("trace_id   :", span.SpanContext().TraceID(), "(32 hex)")
	fmt.Println("span_id    :", span.SpanContext().SpanID(), "(16 hex)")
	fmt.Println("amostrado  :", span.SpanContext().IsSampled())

	// o servico recebe e continua o mesmo trace
	ctx2 := propagation.TraceContext{}.Extract(context.Background(), propagation.HeaderCarrier(h))
	_, filho := tp.Tracer("outro").Start(ctx2, "op-remota")
	fmt.Println("mesmo trace:", filho.SpanContext().TraceID() == span.SpanContext().TraceID())
	filho.End()

	// atributos de recurso podem vir do ambiente
	r, _ := resource.New(context.Background(), resource.WithFromEnv())
	fmt.Println("recurso do ambiente:", r)
}
