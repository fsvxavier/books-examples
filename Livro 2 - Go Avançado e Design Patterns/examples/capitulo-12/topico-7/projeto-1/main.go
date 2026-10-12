package main

import (
	"errors"
	"fmt"
	"strings"
)

type Est int

const (
	Rascunho Est = iota
	Moderacao
	Publicado
)

func (e Est) String() string { return [...]string{"Rascunho", "Moderacao", "Publicado"}[e] }

type Evento int

const (
	Publicar Evento = iota
	Rejeitar
)

func (e Evento) String() string { return [...]string{"publicar", "rejeitar"}[e] }

var todosEstados = []Est{Rascunho, Moderacao, Publicado}
var todosEventos = []Evento{Publicar, Rejeitar}

// NovoFluxo monta a máquina do documento. admin diz se quem opera é
// administrador (a guarda consulta o valor no momento do evento).
func NovoFluxo(admin func() bool) *Maquina[Est, Evento] {
	return Nova[Est, Evento](Rascunho).
		Permitir(Rascunho, Publicar, Regra[Est]{Para: Moderacao}).
		Permitir(Moderacao, Publicar, Regra[Est]{Para: Publicado, Guarda: admin}).
		Permitir(Moderacao, Rejeitar, Regra[Est]{Para: Rascunho}).
		Permitir(Publicado, Publicar, Regra[Est]{Para: Publicado}).
		Proibir(Rascunho, Rejeitar).
		Proibir(Publicado, Rejeitar)
}

func main() {
	admin := false
	m := NovoFluxo(func() bool { return admin })
	m.Observar(func(c Mudanca[Est, Evento]) {
		fmt.Printf("  [observador] v%d: %v --%v--> %v\n", c.Versao, c.De, c.Evento, c.Para)
	})

	fmt.Println("1) caminho feliz até a moderação")
	fmt.Println("  erro:", m.Enviar(Publicar))

	fmt.Println("2) autor tenta publicar em moderação")
	err := m.Enviar(Publicar)
	fmt.Println("  erro:", err)
	var et *ErroTransicao[Est, Evento]
	fmt.Println("  Is(ErrTransicao):", errors.Is(err, ErrTransicao), "| As:", errors.As(err, &et), "| estado:", et.De, "| motivo:", et.Motivo)

	fmt.Println("3) evento sem regra")
	m2 := NovoFluxo(func() bool { return true })
	fmt.Println("  erro:", m2.Enviar(Rejeitar))

	fmt.Println("4) administrador publica")
	admin = true
	fmt.Println("  erro:", m.Enviar(Publicar), "| estado:", m.Estado())

	fmt.Println("5) lacunas (nenhuma, tudo permitido ou proibido de forma explícita)")
	fmt.Println(" ", len(m.Lacunas(todosEstados, todosEventos)), "lacunas")

	fmt.Println("6) esquecendo uma decisão")
	incompleto := Nova[Est, Evento](Rascunho).
		Permitir(Rascunho, Publicar, Regra[Est]{Para: Moderacao}).
		Permitir(Moderacao, Publicar, Regra[Est]{Para: Publicado})
	l := incompleto.Lacunas(todosEstados, todosEventos)
	fmt.Printf("  %d lacunas: %s\n", len(l), strings.Join(l, "; "))

	fmt.Println("7) ação que falha impede a transição")
	m3 := Nova[Est, Evento](Rascunho).Permitir(Rascunho, Publicar, Regra[Est]{
		Para: Moderacao,
		Acao: func() error { return errors.New("serviço de moderação indisponível") },
	})
	fmt.Println("  erro:", m3.Enviar(Publicar), "| estado continua:", m3.Estado())

	fmt.Println("8) diagrama")
	fmt.Println(m.Mermaid())
}
