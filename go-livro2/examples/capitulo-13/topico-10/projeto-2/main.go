package main

import (
	"errors"
	"fmt"
)

type No interface{ isNo() } // interface selada: só este pacote cria nós

type Cidade struct{ Nome string }
type Industria struct{ Nome string }
type PontoTuristico struct{ Nome string }

// Porto é o elemento NOVO. Ninguém atualizou os visitantes.
type Porto struct{ Nome string }

func (*Cidade) isNo()         {}
func (*Industria) isNo()      {}
func (*PontoTuristico) isNo() {}
func (*Porto) isNo()          {}

// ErrNaoTratado é o que um visitante por type switch deve devolver no
// default, em vez de ignorar o elemento em silêncio.
var ErrNaoTratado = errors.New("tipo de nó não tratado")

// Silencioso: o default engole o que não conhece.
func ContarSilencioso(nos []No) (cidades int) {
	for _, n := range nos {
		switch n.(type) {
		case *Cidade:
			cidades++
		default:
			// "não é cidade, segue": e se for um tipo que ESQUECEMOS?
		}
	}
	return cidades
}

// Descrever lista todos os tipos e FALHA no default.
func Descrever(n No) (string, error) {
	switch n := n.(type) {
	case *Cidade:
		return "cidade " + n.Nome, nil
	case *Industria:
		return "indústria " + n.Nome, nil
	case *PontoTuristico:
		return "ponto " + n.Nome, nil
	default:
		return "", fmt.Errorf("%w: %T", ErrNaoTratado, n)
	}
}

// Amostras devolve UM exemplar de cada tipo de nó. É a lista que o teste
// de exaustividade percorre: quem cria um tipo novo acrescenta aqui.
func Amostras() []No {
	return []No{&Cidade{"c"}, &Industria{"i"}, &PontoTuristico{"p"}, &Porto{"x"}}
}

func main() {
	nos := Amostras()
	fmt.Println("cidades (silencioso):", ContarSilencioso(nos))
	for _, n := range nos {
		s, err := Descrever(n)
		fmt.Printf("  %-18q erro: %v\n", s, err)
	}
}
