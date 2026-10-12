package main

import (
	"fmt"
	"time"
)

// ---- 1. Os literais do desconto, com float64 de dinheiro. ----

func descontoRuim(valorTotal float64) float64 {
	if valorTotal > 1000 {
		return valorTotal * 0.1
	}
	return valorTotal * 0.05
}

// Centavos inteiros e constantes com nome: intenção explícita, aritmética exata.
const (
	limiarPremiumCentavos = 100_000 // R$ 1.000,00
	percentualPremium     = 10      // por cento
	percentualPadrao      = 5
)

func desconto(centavos int64) int64 {
	if centavos > limiarPremiumCentavos {
		return centavos * percentualPremium / 100
	}
	return centavos * percentualPadrao / 100
}

// ---- 2. Duração sem tipo: o clássico time.Sleep(5). ----

type Config struct {
	TimeoutSegundos int
	Espera          time.Duration
}

// ---- 3. Enum com iota: o valor zero é um estado válido por acidente. ----

type StatusRuim int

const (
	Pendente StatusRuim = iota // 0: qualquer struct não inicializada "está pendente"
	Pago
	Cancelado
)

type Status int

const (
	StatusDesconhecido Status = iota // 0 reservado para "não informado"
	StatusPendente
	StatusPago
	StatusCancelado
)

func (s Status) String() string {
	switch s {
	case StatusPendente:
		return "pendente"
	case StatusPago:
		return "pago"
	case StatusCancelado:
		return "cancelado"
	}
	return fmt.Sprintf("Status(%d)", int(s))
}

type PedidoRuim struct{ Status StatusRuim }
type Pedido struct{ Status Status }

func main() {
	fmt.Println("== float64 para dinheiro")
	a, b := 0.1, 0.2
	fmt.Println("constantes: 0.1 + 0.2 =", 0.1+0.2, "| variáveis: a + b =", a+b)
	var soma float64
	for range 1_000_000 {
		soma += 0.10
	}
	fmt.Printf("um milhão de centavos de R$ 0,10 somados: %.6f (esperado 100000.000000)\n", soma)
	fmt.Println("descontoRuim(1000.00) =", descontoRuim(1000.00), "(não passa de > 1000)")
	fmt.Println("descontoRuim(1000.01) =", descontoRuim(1000.01))
	preco := 19.99
	fmt.Println("19.99 * 3 (variável) =", preco*3, "| constante:", 19.99*3)
	fmt.Println("desconto(100000) centavos =", desconto(100_000), "| desconto(100001) =", desconto(100_001))

	fmt.Println("== duração")
	t0 := time.Now()
	time.Sleep(5) // 5 nanossegundos, não 5 segundos
	fmt.Printf("time.Sleep(5) levou %v\n", time.Since(t0).Round(time.Microsecond))

	c := Config{TimeoutSegundos: 5, Espera: 5 * time.Second}
	fmt.Println("Espera como Duration:", c.Espera)
	fmt.Println("TimeoutSegundos como int, convertido:", time.Duration(c.TimeoutSegundos)*time.Second)
	esperaBug := c.Espera * time.Second // Duration * Duration compila e dá lixo
	fmt.Println("Espera * time.Second =", esperaBug, "=", esperaBug.Hours(), "horas")

	fmt.Println("== enum")
	var pr PedidoRuim
	var p Pedido
	fmt.Println("PedidoRuim zerado está pendente?", pr.Status == Pendente)
	fmt.Println("Pedido zerado:", p.Status, "| precisa ser validado antes de salvar")
	fmt.Println("Status(7):", Status(7))
}
