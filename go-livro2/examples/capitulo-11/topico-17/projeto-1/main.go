package main

import (
	"fmt"
	"sync"
)

func reais(c int) string { return fmt.Sprintf("R$ %d,%02d", c/100, c%100) }

// Painel mostra o total mais recente. Ignora eventos velhos pela Versao.
type Painel struct {
	mu     sync.Mutex
	versao uint64
	total  int
}

func (p *Painel) Atualizar(m Mudanca) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if m.Versao <= p.versao {
		return // evento atrasado: já vimos um mais novo
	}
	p.versao, p.total = m.Versao, m.Total
}

func (p *Painel) Total() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.total
}

// Alerta avisa na transição, não a cada evento.
type Alerta struct {
	limite int
	acima  bool
}

func (a *Alerta) Atualizar(m Mudanca) {
	if estourou := m.Total > a.limite; estourou != a.acima {
		a.acima = estourou
		if estourou {
			fmt.Printf("   ALERTA: %s passou do limite de %s\n", reais(m.Total), reais(a.limite))
		} else {
			fmt.Printf("   ALERTA: voltou para %s, abaixo do limite\n", reais(m.Total))
		}
	}
}

func main() {
	o := &Orcamento{}
	s := NovaSessao(o)
	painel := &Painel{}
	alerta := &Alerta{limite: 100000}
	var auditoria []Mudanca

	o.Observar(painel.Atualizar)
	o.Observar(alerta.Atualizar)
	cancelaAud := o.Observar(func(m Mudanca) { auditoria = append(auditoria, m) })

	fmt.Println("1) edições normais")
	_ = s.Adicionar("notebook", 450000)
	_ = s.Adicionar("mouse", 8000)
	fmt.Println("   painel:", reais(painel.Total()))

	fmt.Println("2) desfazer e refazer avisam os observadores")
	s.Desfazer()
	fmt.Println("   painel depois de desfazer:", reais(painel.Total()))
	s.Refazer()
	fmt.Println("   painel depois de refazer: ", reais(painel.Total()))

	fmt.Println("3) restaurar em silêncio deixa o painel velho")
	foto := o.Salvar()
	_ = s.Remover("notebook")
	fmt.Println("   painel:", reais(painel.Total()), "| real:", reais(o.Total()))
	o.RestaurarEmSilencio(foto)
	fmt.Println("   painel:", reais(painel.Total()), "| real:", reais(o.Total()), "<- divergiu")
	o.Restaurar(foto)
	fmt.Println("   com Restaurar: painel:", reais(painel.Total()), "| real:", reais(o.Total()))

	fmt.Println("4) lote: um único evento")
	antes := len(auditoria)
	err := o.Lote(
		func(it *[]Item) error { *it = append(*it, Item{"teclado", 30000}); return nil },
		func(it *[]Item) error { *it = append(*it, Item{"monitor", 90000}); return nil },
	)
	fmt.Println("   erro:", err, "| eventos novos:", len(auditoria)-antes)

	fmt.Println("5) lote que falha: nada muda, nada é publicado")
	antes = len(auditoria)
	err = o.Lote(
		func(it *[]Item) error { *it = append(*it, Item{"cadeira", 70000}); return nil },
		func(it *[]Item) error { return fmt.Errorf("fornecedor fora do ar") },
	)
	fmt.Println("   erro:", err)
	fmt.Println("   eventos novos:", len(auditoria)-antes, "| total:", reais(o.Total()))

	cancelaAud()
	_ = s.Adicionar("fone", 1000)
	fmt.Println("6) após cancelar a auditoria, ela parou em", len(auditoria), "eventos")
	for _, m := range auditoria[:3] {
		fmt.Printf("   v%d %-10s itens=%d total=%s\n", m.Versao, m.Tipo, m.Itens, reais(m.Total))
	}
}
