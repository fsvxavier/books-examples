package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func reais(c int) string { return fmt.Sprintf("R$ %d,%02d", c/100, c%100) }

type Visao struct {
	ID     string `json:"id"`
	Estado Est    `json:"estado"`
	Total  int    `json:"total"`
}

func main() {
	itens := []Item{{"livro", 6000, 700}, {"fone", 12000, 300}}

	fmt.Println("1) o cliente escolhe a estratégia de frete")
	for _, c := range []struct {
		nome string
		f    Frete
	}{
		{"fixo", FreteFixo},
		{"por quilo", FretePorKg},
		{"grátis acima de R$ 150", FreteGratis(15000)},
	} {
		p, _ := NovoPedido("P", c.f, itens...)
		fmt.Printf("   %-24s subtotal %s + frete %s = %s\n", c.nome, reais(p.Subtotal()), reais(p.Frete()), reais(p.Total()))
	}

	fmt.Println("2) ciclo de vida, com um observador")
	p, _ := NovoPedido("P-1", FreteGratis(15000), itens...)
	p.Observar(func(m Mudanca[Est, Evento]) {
		fmt.Printf("   [auditoria] v%d %s: %s -> %s\n", m.Versao, m.Evento, m.De, m.Para)
	})
	for _, f := range []func() error{p.Pagar, p.Enviar} {
		_ = f()
	}
	err := p.Cancelar()
	fmt.Println("   cancelar enviado:", err, "| Is:", errors.Is(err, ErrTransicao))
	_ = p.Entregar()

	fmt.Println("3) o ESTADO escolhe a política de cancelamento")
	a, _ := NovoPedido("A", FreteFixo, itens...)
	_ = a.Cancelar()
	fmt.Printf("   cancelado em criado: reembolso %s\n", reais(a.Reembolso()))
	b, _ := NovoPedido("B", FreteFixo, itens...)
	_ = b.Pagar()
	_ = b.Cancelar()
	fmt.Printf("   cancelado em pago:   reembolso %s (total %s, retém 10%%)\n", reais(b.Reembolso()), reais(b.Total()))

	fmt.Println("4) persistir o estado como texto e reidratar")
	pago, _ := NovoPedido("P-2", FreteFixo, itens...)
	_ = pago.Pagar()
	bs, _ := json.Marshal(Visao{pago.ID, pago.Estado(), pago.Total()})
	fmt.Println("   json:", string(bs))
	var v Visao
	_ = json.Unmarshal(bs, &v)
	r, _ := reidratar(v.ID, v.Estado, FreteFixo, itens)
	fmt.Println("   reidratado em:", r.Estado(), "| enviar:", r.Enviar(), "| agora:", r.Estado())
	err = json.Unmarshal([]byte(`{"id":"X","estado":"perdido"}`), &v)
	fmt.Println("   estado inválido:", err)

	fmt.Println("5) completude e diagrama")
	fmt.Println("   lacunas:", len(p.maq.Lacunas(todosEstados, todosEventos)))
	incompleto, _ := NovoPedido("I", FreteFixo, itens...)
	delete(incompleto.maq.proibidos, par[Est, Evento]{Enviado, EvCancelar})
	fmt.Println("   sem decidir enviado+cancelar:", incompleto.maq.Lacunas(todosEstados, todosEventos))
	fmt.Println(strings.ReplaceAll(p.maq.Mermaid(), "\n", "\n   "))
}
