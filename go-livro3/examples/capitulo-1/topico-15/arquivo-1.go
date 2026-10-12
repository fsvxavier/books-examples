// Exemplo à parte: encaixar um modelo próprio e um retriever próprio nas
// cadeias do LangChainGo. Tudo falso, nada de rede.
package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// ModeloFalso satisfaz llms.Model: GenerateContent e Call.
type ModeloFalso struct{ Pedidos []string }

var reContexto = regexp.MustCompile(`(?s)Use the following pieces of context.*?\n\n(.*?)\n\nQuestion:`)

func (m *ModeloFalso) GenerateContent(ctx context.Context, msgs []llms.MessageContent, opts ...llms.CallOption) (*llms.ContentResponse, error) {
	var b strings.Builder
	for _, msg := range msgs {
		for _, p := range msg.Parts {
			if t, ok := p.(llms.TextContent); ok {
				b.WriteString(t.Text)
			}
		}
	}
	prompt := b.String()
	m.Pedidos = append(m.Pedidos, prompt)
	resp := "(modelo falso) recebi um prompt de " + fmt.Sprint(len(prompt)) + " caracteres"
	if sub := reContexto.FindStringSubmatch(prompt); sub != nil {
		resp = "(modelo falso) li o contexto: " + strings.SplitN(sub[1], "\n", 2)[0]
	}
	// Se o chamador pediu streaming, entrega em pedaços.
	var o llms.CallOptions
	for _, f := range opts {
		f(&o)
	}
	if o.StreamingFunc != nil {
		for _, w := range strings.SplitAfter(resp, " ") {
			if err := o.StreamingFunc(ctx, []byte(w)); err != nil {
				return nil, err
			}
		}
	}
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: resp}}}, nil
}

func (m *ModeloFalso) Call(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, m, prompt, opts...)
}

// Retriever satisfaz schema.Retriever.
type Retriever struct{ Docs []string }

func (r Retriever) GetRelevantDocuments(ctx context.Context, q string) ([]schema.Document, error) {
	var out []schema.Document
	for _, d := range r.Docs {
		for _, w := range strings.Fields(strings.ToLower(q)) {
			if len(w) > 3 && strings.Contains(strings.ToLower(d), strings.Trim(w, "?.,")) {
				out = append(out, schema.Document{PageContent: d, Metadata: map[string]any{"fonte": "faq"}})
				break
			}
		}
	}
	return out, nil
}

func main() {
	ctx := context.Background()
	modelo := &ModeloFalso{}
	ret := Retriever{Docs: []string{
		"A rede wifi da biblioteca é gratuita e a senha é pedida no balcão.",
		"A biblioteca abre de segunda a sexta, das 8h às 20h.",
	}}
	cadeia := chains.NewRetrievalQAFromLLM(modelo, ret)
	resp, err := chains.Run(ctx, cadeia, "Qual a senha do wifi?")
	if err != nil {
		panic(err)
	}
	fmt.Println("resposta:", resp)
	fmt.Println("\nprompt que o LangChainGo montou:")
	fmt.Println(modelo.Pedidos[0])
}
