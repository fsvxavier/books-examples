package main

import (
	"encoding/json"
	"sync"
	"testing"
)

type tela struct{ n int }

func (t *tela) Pintar(string, string, int, float64, float64) { t.n++ }

var tex = make([]byte, 64)

func TestCompartilhamento(t *testing.T) {
	f := NovaFabricaTipos()
	fl := NovaFloresta(f)
	for i := range 1000 {
		fl.Plantar(float64(i), 0, "carvalho", "verde", tex)
	}
	if f.Quantidade() != 1 {
		t.Fatalf("flyweights = %d", f.Quantidade())
	}
	for _, a := range fl.arvores {
		if a.Tipo != fl.arvores[0].Tipo {
			t.Fatal("arvores do mesmo tipo com ponteiros diferentes")
		}
	}
}

func TestFabricaCopiaATextura(t *testing.T) {
	f := NovaFabricaTipos()
	origem := []byte{1, 2, 3}
	tipo := f.Obter("a", "b", origem)
	origem[0] = 99 // o chamador mexe no slice dele
	if tipo.textura[0] != 1 {
		t.Fatal("o flyweight guardou uma referencia ao slice do chamador")
	}
}

func TestDesenharChamaUmaVezPorArvore(t *testing.T) {
	fl := NovaFloresta(NovaFabricaTipos())
	for range 10 {
		fl.Plantar(1, 2, "x", "y", tex)
	}
	var tl tela
	fl.Desenhar(&tl)
	if tl.n != 10 {
		t.Fatalf("pinturas = %d", tl.n)
	}
}

func TestIdaEVoltaPreservaContextoECompartilhamento(t *testing.T) {
	cat := map[string]TipoDef{"a": {"verde", tex}, "b": {"azul", tex}}
	f := NovaFabricaTipos()
	fl := NovaFloresta(f)
	fl.Plantar(1.5, 2.5, "a", "verde", tex)
	fl.Plantar(3, 4, "b", "azul", tex)
	fl.Plantar(5, 6, "a", "verde", tex)
	dados, err := json.Marshal(fl)
	if err != nil {
		t.Fatal(err)
	}
	f2 := NovaFabricaTipos()
	fl2, err := Carregar(dados, f2, cat)
	if err != nil {
		t.Fatal(err)
	}
	if fl2.Tamanho() != 3 || f2.Quantidade() != 2 || fl2.arvores[0].Tipo != fl2.arvores[2].Tipo {
		t.Fatalf("tamanho=%d flyweights=%d", fl2.Tamanho(), f2.Quantidade())
	}
	if fl2.arvores[0].X != 1.5 || fl2.arvores[1].Y != 4 {
		t.Fatalf("contexto perdido: %+v", fl2.arvores)
	}
}

func TestTipoDesconhecido(t *testing.T) {
	_, err := Carregar([]byte(`[{"x":1,"y":2,"tipo":"z"}]`), NovaFabricaTipos(), map[string]TipoDef{})
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestPlantioConcorrente(t *testing.T) {
	f := NovaFabricaTipos()
	var wg sync.WaitGroup
	nomes := []string{"a", "b", "c", "d"}
	var mu sync.Mutex
	vistos := map[string]map[*TipoArvore]bool{}
	for i := range 400 {
		wg.Go(func() {
			n := nomes[i%4]
			tp := f.Obter(n, "cor", tex)
			mu.Lock()
			if vistos[n] == nil {
				vistos[n] = map[*TipoArvore]bool{}
			}
			vistos[n][tp] = true
			mu.Unlock()
		})
	}
	wg.Wait()
	if f.Quantidade() != 4 {
		t.Fatalf("flyweights = %d", f.Quantidade())
	}
	for n, m := range vistos {
		if len(m) != 1 {
			t.Fatalf("%s: %d instancias", n, len(m))
		}
	}
}
