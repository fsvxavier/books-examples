package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
)

type chave struct{}

func main() {
	// ---- http.Request.Clone: copia PROFUNDA dos cabecalhos, RASA do corpo ----
	base, _ := http.NewRequest("POST", "https://api.exemplo.com/pedidos", strings.NewReader("corpo"))
	base.Header.Set("Authorization", "Bearer abc")

	dup := base.Clone(context.WithValue(context.Background(), chave{}, "req-2"))
	dup.Header.Set("Authorization", "Bearer outro")
	dup.Header.Add("X-Extra", "1")
	fmt.Println("Request.Clone:")
	fmt.Println("  original Authorization:", base.Header.Get("Authorization"), "| X-Extra:", base.Header.Get("X-Extra") == "")
	fmt.Println("  clone    Authorization:", dup.Header.Get("Authorization"))
	fmt.Println("  corpo compartilhado?  ", base.Body == dup.Body, "<- o Body (um leitor) NAO e copiado")
	fmt.Println("  contexto do clone:    ", dup.Context().Value(chave{}))

	// ---- WithContext: copia RASA (cabecalhos compartilhados!) ----
	raso := base.WithContext(context.Background())
	raso.Header.Set("X-Compartilhado", "sim")
	fmt.Println("WithContext (rasa): original viu o cabecalho novo?", base.Header.Get("X-Compartilhado") == "sim")

	// ---- http.Header.Clone ----
	h := http.Header{"A": {"1"}}
	hc := h.Clone()
	hc.Add("A", "2")
	fmt.Println("Header.Clone:", h["A"], "contra", hc["A"])

	// ---- tls.Config.Clone: o exemplo da biblioteca padrao ----
	cfg := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: "api.exemplo.com"}
	c2 := cfg.Clone()
	c2.MinVersion = tls.VersionTLS13
	fmt.Printf("tls.Config.Clone: original=%s clone=%s\n", tls.VersionName(cfg.MinVersion), tls.VersionName(c2.MinVersion))

	// ---- http.Transport.Clone: base para clientes derivados ----
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConnsPerHost = 50
	fmt.Println("Transport.Clone: padrao =", http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost, "| derivado =", t.MaxIdleConnsPerHost)
}
