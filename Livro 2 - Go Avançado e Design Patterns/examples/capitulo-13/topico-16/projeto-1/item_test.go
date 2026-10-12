package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// Exaustividade: todo visitante precisa tratar todo tipo de item.
func TestVisitantesTratamTodosOsTipos(t *testing.T) {
	for _, it := range Amostras() {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%T: %v", it, r)
				}
			}()
			Aceitar[int](it, Valor{})
			Aceitar[int](it, Imposto{})
			Aceitar[string](it, Descricao{})
		}()
	}
}

func TestImpostos(t *testing.T) {
	casos := []struct {
		item Item
		want int
	}{
		{Livro{"x", 10000}, 0},
		{Eletronico{"x", 10000, 12}, 1800},
		{Alimento{"x", 10000, false}, 700},
		{Alimento{"x", 10000, true}, 400},
		{Servico{"x", 2, 5000}, 500},
	}
	for _, c := range casos {
		if got := Aceitar[int](c.item, Imposto{}); got != c.want {
			t.Errorf("%T: imposto %d, quero %d", c.item, got, c.want)
		}
	}
}

func TestTotaisIguaisEmTodosOsFormatos(t *testing.T) {
	pedido := []Item{Livro{"a", 12000}, Eletronico{"b", 45000, 12}, Servico{"c", 3, 9000}}
	var ref Totais
	for i, f := range []Formato{Texto{}, CSV{}, HTML{}} {
		var buf bytes.Buffer
		tot, err := Emitir(&buf, f, pedido)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			ref = tot
		} else if tot != ref {
			t.Fatalf("%T: %v != %v", f, tot, ref)
		}
	}
	if ref.Valor != 84000 || ref.Imposto != 8100+1350 {
		t.Fatalf("totais %+v", ref)
	}
}

func TestHTMLEscapa(t *testing.T) {
	var buf bytes.Buffer
	_, _ = Emitir(&buf, HTML{}, []Item{Eletronico{"<script>", 100, 1}})
	if strings.Contains(buf.String(), "<script>") {
		t.Fatalf("HTML não escapou: %s", buf.String())
	}
}

type escritaQueFalha struct{ restam int }

func (e *escritaQueFalha) Write(p []byte) (int, error) {
	if e.restam == 0 {
		return 0, errors.New("disco cheio")
	}
	e.restam--
	return len(p), nil
}

func TestErroDeEscritaInterrompe(t *testing.T) {
	_, err := Emitir(&escritaQueFalha{restam: 2}, Texto{}, Amostras())
	if err == nil || err.Error() != "disco cheio" {
		t.Fatalf("err = %v", err)
	}
	_, err = Emitir(io.Discard, Texto{}, Amostras())
	if err != nil {
		t.Fatal(err)
	}
}

func TestCSVEscapaSeparadorEAspas(t *testing.T) {
	var buf bytes.Buffer
	_, _ = Emitir(&buf, CSV{}, []Item{Livro{`Go; "a linguagem"`, 100}})
	want := "item;valor;imposto\n\"Livro: Go; \"\"a linguagem\"\"\";1,00;0,00\ntotal;1,00;0,00\n"
	if buf.String() != want {
		t.Fatalf("got %q\nwant %q", buf.String(), want)
	}
}
