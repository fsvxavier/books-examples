package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/attribute"
	otmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// TraceHandler acrescenta trace_id e span_id do span ativo no contexto.
type TraceHandler struct{ slog.Handler }

func (h TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(slog.String("trace_id", sc.TraceID().String()), slog.String("span_id", sc.SpanID().String()))
	}
	return h.Handler.Handle(ctx, r)
}
func (h TraceHandler) WithAttrs(a []slog.Attr) slog.Handler {
	return TraceHandler{h.Handler.WithAttrs(a)}
}
func (h TraceHandler) WithGroup(n string) slog.Handler { return TraceHandler{h.Handler.WithGroup(n)} }

func contar(amostrador sdktrace.Sampler, n int) int {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp), sdktrace.WithSampler(amostrador))
	t := tp.Tracer("teste")
	for range n {
		_, s := t.Start(context.Background(), "op")
		s.End()
	}
	return len(exp.GetSpans())
}

func main() {
	// 1. correlacao log <-> trace
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	log := slog.New(TraceHandler{slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(g []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		}})})

	ctx, span := tp.Tracer("pedidos").Start(context.Background(), "criar-pedido")
	log.InfoContext(ctx, "pedido criado", "pedido_id", 42)
	log.Info("sem contexto, sem trace")
	span.End()
	fmt.Println("span.trace_id =", span.SpanContext().TraceID())

	// 2. amostragem
	const n = 10000
	fmt.Println()
	fmt.Println("amostragem de", n, "traces:")
	fmt.Println("  AlwaysSample:                  ", contar(sdktrace.AlwaysSample(), n))
	fmt.Println("  NeverSample:                   ", contar(sdktrace.NeverSample(), n))
	fmt.Println("  TraceIDRatioBased(0.1):        ", contar(sdktrace.TraceIDRatioBased(0.1), n))
	fmt.Println("  ParentBased(TraceIDRatio(0.1)):", contar(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1)), n))

	// 3. metricas do OpenTelemetry com leitor manual
	leitor := metric.NewManualReader()
	mp := metric.NewMeterProvider(metric.WithReader(leitor))
	m := mp.Meter("loja")
	pedidos, _ := m.Int64Counter("pedidos.criados", otmetric.WithDescription("Pedidos criados"))
	dur, _ := m.Float64Histogram("pedido.duracao", otmetric.WithUnit("s"),
		otmetric.WithExplicitBucketBoundaries(0.01, 0.1, 1))
	for _, d := range []float64{0.005, 0.05, 0.05, 0.5, 2} {
		pedidos.Add(ctx, 1, otmetric.WithAttributes(attribute.String("canal", "web")))
		dur.Record(ctx, d)
	}
	pedidos.Add(ctx, 3, otmetric.WithAttributes(attribute.String("canal", "app")))

	var rm metricdata.ResourceMetrics
	leitor.Collect(ctx, &rm)
	fmt.Println()
	for _, sm := range rm.ScopeMetrics {
		for _, met := range sm.Metrics {
			switch d := met.Data.(type) {
			case metricdata.Sum[int64]:
				for _, p := range d.DataPoints {
					canal, _ := p.Attributes.Value("canal")
					fmt.Printf("%s{canal=%s} = %d\n", met.Name, canal.AsString(), p.Value)
				}
			case metricdata.Histogram[float64]:
				for _, p := range d.DataPoints {
					fmt.Printf("%s: contagem=%d soma=%.3f baldes=%v limites=%v\n", met.Name, p.Count, p.Sum, p.BucketCounts, p.Bounds)
				}
			}
		}
	}
}
