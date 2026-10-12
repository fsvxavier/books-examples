// Package feature é o contrato das features: o que existe, como se chama e
// como é calculado. Treino e serviço usam exatamente este código.
package feature

import (
	"math"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-3/capitulo-7/topico-7/projeto-1/internal/janela"
)

// Janelas são as três durações usadas pelas features de comportamento.
var Janelas = [3]time.Duration{5 * time.Minute, time.Hour, 24 * time.Hour}

// Nomes de cada posição de Vetor, na ordem em que o modelo as recebe.
var Nomes = []string{"valor", "tx_5m", "soma_5m", "tx_1h", "soma_1h", "tx_24h", "soma_24h", "razao_media_1h"}

// Comportamento são as agregações da entidade (cartão) numa janela.
type Comportamento [3]janela.Agregado

// Vetor transforma o valor da transação que chega e o comportamento do
// cartão no vetor numérico que o modelo consome. Valores em reais; somas
// e contagens passam por log1p para domar a cauda longa.
func Vetor(centav int64, c Comportamento) []float64 {
	valor := float64(centav) / 100
	media1h := 0.0
	if c[1].Contagem > 0 {
		media1h = float64(c[1].Centav) / 100 / float64(c[1].Contagem)
	}
	return []float64{
		math.Log1p(valor),
		math.Log1p(float64(c[0].Contagem)),
		math.Log1p(float64(c[0].Centav) / 100),
		math.Log1p(float64(c[1].Contagem)),
		math.Log1p(float64(c[1].Centav) / 100),
		math.Log1p(float64(c[2].Contagem)),
		math.Log1p(float64(c[2].Centav) / 100),
		math.Log1p(valor / (media1h + 1)),
	}
}

// Calculadora mantém as três janelas em memória.
type Calculadora struct {
	w [3]janela.Janela
}

func NovaCalculadora() *Calculadora {
	var c Calculadora
	for i, d := range Janelas {
		c.w[i] = janela.NovaBaldes(d)
	}
	return &c
}

func (c *Calculadora) Observar(cartao string, quando time.Time, centav int64) {
	for _, w := range c.w {
		w.Adicionar(cartao, quando, centav)
	}
}

func (c *Calculadora) Comportamento(cartao string, agora time.Time) Comportamento {
	var r Comportamento
	for i, w := range c.w {
		r[i] = w.Consultar(cartao, agora)
	}
	return r
}
