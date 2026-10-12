// Programa que chama o mesmo servidor (laya-serve, local) com três clientes Go
// de terceiros escritos para a API do Jev, sem mudar nada além da URL base.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	kjev "github.com/kataras/jev"
	mjev "github.com/mattn/go-jev"
	wjev "github.com/waygatetech/jev-go"
)

const estado = "Fui cobrado duas vezes em março. Estorne a cobrança duplicada hoje ou cancelo o plano."

func main() {
	base := os.Getenv("DECISOR_URL")
	if base == "" {
		base = "http://127.0.0.1:18690"
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// 1. kataras/jev: exige uma chave não vazia; o servidor local a ignora.
	k, err := kjev.New(kjev.WithAPIKey("qualquer"), kjev.WithBaseURL(base), kjev.WithModel("multilingual"))
	if err != nil {
		fmt.Println("kataras/jev New:", err)
	} else {
		t0 := time.Now()
		r, err := k.Noul(ctx, estado, "O usuário ameaça cancelar o plano?")
		fmt.Printf("kataras/jev  Noul:     P=%.4f err=%v (%v)\n", r.Noul, err, time.Since(t0).Round(time.Millisecond))
		c, err := k.Classify(ctx, estado, "Qual área deve atender?", map[string]any{
			"cobranca": "faturas e estornos", "tecnico": "erros e quedas", "vendas": "preços e propostas", "outro": nil})
		fmt.Printf("kataras/jev  Classify: %s conf=%.4f err=%v\n", c.Choice, c.Confidence, err)
	}

	// 2. mattn/go-jev: WithURL recebe o endpoint inteiro (ou Endpoint("host:porta")).
	m := mjev.NewClient(mjev.WithURL(base+"/v1/systemone"), mjev.WithModel("multilingual"))
	a, err := m.Ask(ctx, estado, mjev.Question{Type: "noul", Instructions: "O usuário ameaça cancelar o plano?"})
	if err != nil {
		fmt.Println("mattn/go-jev Ask:", err)
	} else {
		fmt.Printf("mattn/go-jev Ask:      P=%.4f\n", a.Noul)
	}
	a, err = m.Ask(ctx, estado, mjev.Question{Type: "choice", Instructions: "Qual área deve atender?", Criteria: mjev.Options{
		{Name: "cobranca", Desc: "faturas e estornos"}, {Name: "tecnico", Desc: "erros e quedas"}, {Name: "vendas", Desc: "preços e propostas"}, {Name: "outro"}}})
	if err != nil {
		fmt.Println("mattn/go-jev Ask choice:", err)
	} else {
		fmt.Printf("mattn/go-jev Ask:      %s conf=%.4f\n", a.Choice, a.Confidence)
	}

	// 3. waygatetech/jev-go: BaseURL é um campo público.
	w := wjev.NewClient("qualquer")
	w.BaseURL = base
	resp, err := w.Evaluate(ctx, wjev.Request{Model: "multilingual", State: map[string]any{"text": estado},
		Questions: map[string]wjev.Question{"q": {Type: "noul", Instructions: "O usuário ameaça cancelar o plano?"}}})
	if err != nil {
		fmt.Printf("waygatetech/jev-go: HTTP %d err=%v\n", resp.HTTPStatus, err)
	} else {
		fmt.Printf("waygatetech/jev-go:    P=%.4f modelo=%s\n", *resp.Answers["q"].Noul, resp.Model)
	}
}
