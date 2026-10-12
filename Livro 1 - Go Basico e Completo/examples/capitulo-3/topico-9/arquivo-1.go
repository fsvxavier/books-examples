package main

import "fmt"

// Status representa a etapa do pedido no seu ciclo de vida.
type Status int

const (
	Pendente Status = iota
	Pago
	Enviado
	Entregue
	Cancelado
)

func (s Status) String() string {
	switch s {
	case Pendente:
		return "pendente"
	case Pago:
		return "pago"
	case Enviado:
		return "enviado"
	case Entregue:
		return "entregue"
	case Cancelado:
		return "cancelado"
	default:
		return fmt.Sprintf("Status(%d)", int(s))
	}
}

// Pedido guarda os dados que as regras precisam consultar. Valores em centavos.
type Pedido struct {
	Status         Status
	ValorTotal     int64
	ItensNoEstoque bool
}

// ProximoPasso decide o que fazer com o pedido. A ordem dos cases importa:
// o primeiro verdadeiro vence, então as exceções vêm antes das regras gerais.
func ProximoPasso(p Pedido) string {
	switch {
	case p.Status == Cancelado:
		return "encerrar processo e notificar cliente"
	case p.Status == Pendente && p.ValorTotal <= 0:
		return "corrigir o valor antes de prosseguir"
	case p.Status == Pendente && !p.ItensNoEstoque:
		return "aguardar reposição de estoque"
	case p.Status == Pendente:
		return "processar pagamento"
	case p.Status == Pago:
		return "preparar envio"
	case p.Status == Enviado:
		return "aguardar confirmação de entrega"
	case p.Status == Entregue:
		return "solicitar avaliação"
	default:
		return "status desconhecido, revisar manualmente"
	}
}

// Frete calcula o frete em centavos conforme o valor do pedido e a região.
func Frete(valor int64, regiao string) int64 {
	if valor >= 20000 {
		return 0 // frete grátis a partir de R$ 200,00
	}

	switch regiao {
	case "sul", "sudeste":
		return 1290
	case "centro-oeste", "nordeste":
		return 1990
	case "norte":
		return 2990
	default:
		return 2490
	}
}

func main() {
	pedidos := []Pedido{
		{Pendente, 15000, true},
		{Pendente, 15000, false},
		{Pendente, 0, true},
		{Pago, 8000, true},
		{Enviado, 8000, true},
		{Cancelado, 8000, true},
		{Status(9), 100, true},
	}
	for _, p := range pedidos {
		fmt.Printf("%-10v -> %s\n", p.Status, ProximoPasso(p))
	}

	fmt.Println(Frete(15000, "sudeste"), Frete(15000, "norte"), Frete(25000, "norte"), Frete(5000, "exterior"))
}
