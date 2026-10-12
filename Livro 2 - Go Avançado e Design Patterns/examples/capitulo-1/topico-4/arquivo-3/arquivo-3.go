package main

import (
	"fmt"
	"strings"
	"time"
)

// ---- Camada de alto nivel: regra de negocio. Define o que PRECISA. ----

type Relogio interface{ Agora() time.Time }

type Avisador interface {
	Avisar(destino, msg string) error
}

type Lembrete struct {
	Relogio  Relogio
	Avisador Avisador
}

// Lembrar avisa se faltam menos de 24h para o prazo.
func (l Lembrete) Lembrar(destino string, prazo time.Time) (bool, error) {
	faltam := prazo.Sub(l.Relogio.Agora())
	if faltam <= 0 || faltam >= 24*time.Hour {
		return false, nil
	}
	return true, l.Avisador.Avisar(destino, fmt.Sprintf("faltam %s para o prazo", faltam.Round(time.Minute)))
}

// ---- Camada de baixo nivel: detalhes. Dependem da abstracao, nao o contrario. ----

type relogioReal struct{}

func (relogioReal) Agora() time.Time { return time.Now() }

type relogioFixo struct{ t time.Time }

func (r relogioFixo) Agora() time.Time { return r.t }

type avisoNoLog struct{ saida *strings.Builder }

func (a avisoNoLog) Avisar(destino, msg string) error {
	fmt.Fprintf(a.saida, "[aviso para %s] %s\n", destino, msg)
	return nil
}

// ---- Raiz de composicao: o UNICO lugar que conhece os tipos concretos ----

func montar(ambiente string, saida *strings.Builder) Lembrete {
	var r Relogio = relogioReal{}
	if ambiente == "teste" {
		r = relogioFixo{time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)}
	}
	return Lembrete{Relogio: r, Avisador: avisoNoLog{saida}}
}

func main() {
	var saida strings.Builder
	l := montar("teste", &saida)
	prazo := time.Date(2026, 10, 5, 8, 30, 0, 0, time.UTC)

	enviou, err := l.Lembrar("ana@exemplo.com", prazo)
	fmt.Println("prazo em 20h30:", enviou, err)
	enviou, _ = l.Lembrar("ana@exemplo.com", prazo.Add(48*time.Hour))
	fmt.Println("prazo em 3 dias:", enviou)
	enviou, _ = l.Lembrar("ana@exemplo.com", prazo.Add(-48*time.Hour))
	fmt.Println("prazo vencido:  ", enviou)
	fmt.Print(saida.String())
}
