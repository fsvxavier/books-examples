package main

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
)

// Registro e um registro de fabricas para QUALQUER tipo de produto P e de configuracao C.
type Registro[P, C any] struct {
	rotulo   string
	mu       sync.RWMutex
	fabricas map[string]func(C) (P, error)
}

func NovoRegistro[P, C any](rotulo string) *Registro[P, C] {
	return &Registro[P, C]{rotulo: rotulo, fabricas: map[string]func(C) (P, error){}}
}

var ErrNaoRegistrado = errors.New("nao registrado")

func (r *Registro[P, C]) Registrar(nome string, f func(C) (P, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.fabricas[nome]; dup || f == nil {
		panic(fmt.Sprintf("%s: registro invalido de %q", r.rotulo, nome))
	}
	r.fabricas[nome] = f
}

func (r *Registro[P, C]) Novo(nome string, cfg C) (P, error) {
	r.mu.RLock()
	f, ok := r.fabricas[nome]
	r.mu.RUnlock()
	if !ok {
		var zero P
		return zero, fmt.Errorf("%s %w: %q (opcoes: %s)", r.rotulo, ErrNaoRegistrado, nome, strings.Join(r.Nomes(), ", "))
	}
	return f(cfg)
}

func (r *Registro[P, C]) Nomes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Sorted(maps.Keys(r.fabricas))
}

// ---- dois usos, com tipos diferentes, sem repetir o codigo do registro ----

type Compactador interface{ Compactar(s string) string }

type CfgCompactador struct{ Nivel int }

type Relogio interface{ Hora() string }

var compactadores = NovoRegistro[Compactador, CfgCompactador]("compactador")
var relogios = NovoRegistro[Relogio, string]("relogio") // a configuracao aqui e um fuso (string)

type espacos struct{}

func (espacos) Compactar(s string) string { return strings.Join(strings.Fields(s), " ") }

type vogais struct{ nivel int }

func (v vogais) Compactar(s string) string {
	return strings.NewReplacer("a", "", "e", "", "i", "", "o", "", "u", "").Replace(s)
}

type fixo struct{ fuso string }

func (f fixo) Hora() string { return "12:00 em " + f.fuso }

func main() {
	compactadores.Registrar("espacos", func(CfgCompactador) (Compactador, error) { return espacos{}, nil })
	compactadores.Registrar("vogais", func(c CfgCompactador) (Compactador, error) {
		if c.Nivel < 1 {
			return nil, errors.New("vogais: nivel deve ser >= 1")
		}
		return vogais{c.Nivel}, nil
	})
	relogios.Registrar("fixo", func(fuso string) (Relogio, error) { return fixo{fuso}, nil })

	fmt.Println("compactadores:", compactadores.Nomes())
	c, _ := compactadores.Novo("espacos", CfgCompactador{})
	fmt.Printf("%q\n", c.Compactar("  muitos    espacos   aqui "))
	c, _ = compactadores.Novo("vogais", CfgCompactador{Nivel: 1})
	fmt.Printf("%q\n", c.Compactar("registro generico"))

	_, err := compactadores.Novo("vogais", CfgCompactador{})
	fmt.Println("config invalida:", err)
	_, err = compactadores.Novo("zip", CfgCompactador{})
	fmt.Println("desconhecido:   ", err, "| errors.Is:", errors.Is(err, ErrNaoRegistrado))

	r, _ := relogios.Novo("fixo", "America/Sao_Paulo")
	fmt.Println(r.Hora())
	_, err = relogios.Novo("atomico", "")
	fmt.Println("outro registro: ", err)

	// o compilador impede misturar: relogios.Novo("fixo", CfgCompactador{}) nao compila
}
