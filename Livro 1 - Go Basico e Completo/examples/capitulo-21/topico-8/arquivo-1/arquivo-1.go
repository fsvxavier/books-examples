package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

var tracer = otel.Tracer("loja/pedidos")

func buscarEstoque(ctx context.Context, sku string) (int, error) {
	_, span := tracer.Start(ctx, "estoque.buscar")
	defer span.End()
	span.SetAttributes(attribute.String("sku", sku))
	time.Sleep(5 * time.Millisecond)
	if sku == "RUIM" {
		err := errors.New("banco indisponivel")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return 0, err
	}
	return 10, nil
}

func main() {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp)) // WithSyncer: sem lote, ideal em teste
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	// servidor: o otelhttp cria o span da requisicao e le o traceparent do cabecalho
	mux := http.NewServeMux()
	mux.HandleFunc("GET /estoque/{sku}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("servidor recebeu traceparent:", r.Header.Get("Traceparent") != "")
		n, err := buscarEstoque(r.Context(), r.PathValue("sku"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, n)
	})
	srv := httptest.NewServer(otelhttp.NewHandler(mux, "servidor"))
	defer srv.Close()

	// cliente: o Transport do otelhttp cria um span e INJETA o traceparent
	cliente := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	ctx, raiz := tracer.Start(context.Background(), "fazer-pedido")
	for _, sku := range []string{"A1", "RUIM"} {
		req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL+"/estoque/"+sku, nil)
		resp, err := cliente.Do(req)
		if err != nil {
			panic(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		fmt.Println("resposta", sku, "->", resp.StatusCode)
	}
	raiz.End()
	tp.ForceFlush(ctx)

	spans := exp.GetSpans()
	sort.Slice(spans, func(i, j int) bool { return spans[i].StartTime.Before(spans[j].StartTime) })

	nome := map[string]string{}
	for _, s := range spans {
		nome[s.SpanContext.SpanID().String()] = s.Name
	}
	fmt.Println()
	trace0 := spans[0].SpanContext.TraceID()
	mesmo := true
	for _, s := range spans {
		if s.SpanContext.TraceID() != trace0 {
			mesmo = false
		}
		pai := "(raiz)"
		if s.Parent.IsValid() {
			pai = nome[s.Parent.SpanID().String()]
		}
		fmt.Printf("%-18s pai=%-18s tipo=%-8s status=%s\n", s.Name, pai, s.SpanKind, s.Status.Code)
	}
	fmt.Println("todos no mesmo trace:", mesmo)
}
