package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

// ---- logs: handler que injeta trace_id e span_id ----

type TraceHandler struct{ slog.Handler }

func (h TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(slog.String("trace_id", sc.TraceID().String()))
	}
	return h.Handler.Handle(ctx, r)
}
func (h TraceHandler) WithAttrs(a []slog.Attr) slog.Handler {
	return TraceHandler{h.Handler.WithAttrs(a)}
}
func (h TraceHandler) WithGroup(n string) slog.Handler { return TraceHandler{h.Handler.WithGroup(n)} }

// ---- metricas ----

var (
	reqs = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total", Help: "Requisicoes HTTP."}, []string{"rota", "status"})
	dur = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds", Help: "Duracao.", Buckets: []float64{.01, .05, .1, .5}}, []string{"rota"})
)

type gravador struct {
	http.ResponseWriter
	status int
}

func (g *gravador) WriteHeader(c int)           { g.status = c; g.ResponseWriter.WriteHeader(c) }
func (g *gravador) Unwrap() http.ResponseWriter { return g.ResponseWriter }

func medir(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g := &gravador{ResponseWriter: w, status: 200}
		ini := time.Now()
		next.ServeHTTP(g, r)
		d := time.Since(ini)
		reqs.WithLabelValues(r.Pattern, fmt.Sprint(g.status)).Inc()
		dur.WithLabelValues(r.Pattern).Observe(d.Seconds())
		nivel := slog.LevelInfo
		if g.status >= 500 {
			nivel = slog.LevelError
		}
		log.Log(r.Context(), nivel, "requisicao", "rota", r.Pattern, "status", g.status, "ms", d.Milliseconds())
	})
}

func main() {
	// um "coletor" OTLP de mentira, para provar que o exportador envia mesmo
	var recebidos atomic.Int64
	var bytesRecebidos atomic.Int64
	var caminho atomic.Value
	coletor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		recebidos.Add(1)
		bytesRecebidos.Add(int64(len(b)))
		caminho.Store(r.URL.Path + " " + r.Header.Get("Content-Type"))
	}))
	defer coletor.Close()

	ctx := context.Background()
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(coletor.URL+"/v1/traces"),
		otlptracehttp.WithInsecure())
	if err != nil {
		panic(err)
	}
	res := resource.NewSchemaless(
		semconv.ServiceName("loja-api"),
		semconv.ServiceVersion("1.4.0"),
		attribute.String("deployment.environment.name", "lab"),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp, sdktrace.WithBatchTimeout(100*time.Millisecond)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	reg := prometheus.NewRegistry()
	reg.MustRegister(reqs, dur)

	log := slog.New(TraceHandler{slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(g []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		}})})

	tracer := otel.Tracer("loja")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /pedidos/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "consultar-banco")
		time.Sleep(15 * time.Millisecond)
		span.End()
		if r.PathValue("id") == "13" {
			log.ErrorContext(ctx, "falha ao ler pedido", "pedido_id", 13)
			http.Error(w, "erro interno", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "ok")
	})
	mux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	// ordem: otelhttp (cria o span) por fora, para o log e as metricas verem o trace
	srv := httptest.NewServer(otelhttp.NewHandler(medir(log, mux), "http"))

	for _, id := range []string{"1", "2", "13"} {
		resp, _ := http.Get(srv.URL + "/pedidos/" + id)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	resp, _ := http.Get(srv.URL + "/metrics")
	corpo, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println()
	for _, l := range strings.Split(string(corpo), "\n") {
		if strings.HasPrefix(l, "http_requests_total") {
			fmt.Println(l)
		}
	}

	// encerramento: Shutdown do provider esvazia o lote pendente
	srv.Close()
	ctxFim, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := tp.Shutdown(ctxFim); err != nil {
		fmt.Println("shutdown:", err)
	}
	fmt.Printf("\ncoletor recebeu %d envio(s), %d bytes, em %v\n", recebidos.Load(), bytesRecebidos.Load(), caminho.Load())
}
