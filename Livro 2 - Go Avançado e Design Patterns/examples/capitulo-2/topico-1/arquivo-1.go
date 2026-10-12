package main

import (
	"errors"
	"fmt"
	"strings"
)

// Produto: o que todas as fabricas entregam.
type Transporte interface {
	Entregar(carga string) string
	CustoPorKm() float64
}

type Caminhao struct{}

func (Caminhao) Entregar(c string) string { return "caminhao leva " + c + " por estrada" }
func (Caminhao) CustoPorKm() float64      { return 2.5 }

type Navio struct{}

func (Navio) Entregar(c string) string { return "navio leva " + c + " por mar" }
func (Navio) CustoPorKm() float64      { return 0.8 }

type Aviao struct{ Capacidade int }

func (a Aviao) Entregar(c string) string {
	return fmt.Sprintf("aviao (%d t) leva %s pelo ar", a.Capacidade, c)
}
func (Aviao) CustoPorKm() float64 { return 12 }

// ---- 1. A forma "classica": um Criador com um metodo-fabrica ----

type Logistica interface {
	CriarTransporte() Transporte // o factory method
}

type LogisticaRodoviaria struct{}

func (LogisticaRodoviaria) CriarTransporte() Transporte { return Caminhao{} }

type LogisticaMaritima struct{}

func (LogisticaMaritima) CriarTransporte() Transporte { return Navio{} }

// PlanejarEntrega e codigo do CLIENTE: nao conhece Caminhao nem Navio.
func PlanejarEntrega(l Logistica, carga string, km float64) string {
	t := l.CriarTransporte()
	return fmt.Sprintf("%s | custo %.2f", t.Entregar(carga), t.CustoPorKm()*km)
}

// ---- 2. A forma idiomatica: uma FUNCAO que devolve a interface ----

var ErrModalDesconhecido = errors.New("modal desconhecido")

func NovoTransporte(modal string) (Transporte, error) {
	switch strings.ToLower(modal) {
	case "rodoviario":
		return Caminhao{}, nil
	case "maritimo":
		return Navio{}, nil
	case "aereo":
		return Aviao{Capacidade: 40}, nil
	}
	return nil, fmt.Errorf("%w: %q", ErrModalDesconhecido, modal)
}

// ---- 3. O "tipo fabrica" funcional: a fabrica como valor ----

type FabricaDeTransporte func() Transporte

func PlanejarComFabrica(f FabricaDeTransporte, carga string) string { return f().Entregar(carga) }

func main() {
	fmt.Println("1. classica")
	fmt.Println("  ", PlanejarEntrega(LogisticaRodoviaria{}, "geladeiras", 100))
	fmt.Println("  ", PlanejarEntrega(LogisticaMaritima{}, "geladeiras", 100))

	fmt.Println("2. funcao NovoTransporte")
	for _, m := range []string{"rodoviario", "AEREO", "ferroviario"} {
		t, err := NovoTransporte(m)
		if err != nil {
			fmt.Println("   erro:", err, "| errors.Is:", errors.Is(err, ErrModalDesconhecido))
			continue
		}
		fmt.Println("  ", t.Entregar("livros"))
	}

	fmt.Println("3. fabrica como valor")
	fmt.Println("  ", PlanejarComFabrica(func() Transporte { return Navio{} }, "grãos"))
	// um transporte novo, sem tocar em nada que existe:
	drone := FabricaDeTransporte(func() Transporte { return dronePequeno{} })
	fmt.Println("  ", PlanejarComFabrica(drone, "remedios"))
}

type dronePequeno struct{}

func (dronePequeno) Entregar(c string) string { return "drone leva " + c + " ate sua porta" }
func (dronePequeno) CustoPorKm() float64      { return 4 }
