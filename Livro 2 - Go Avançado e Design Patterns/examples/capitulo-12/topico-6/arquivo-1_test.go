package main

import (
	"errors"
	"testing"
)

type fabrica struct {
	nome string
	nova func(admin bool) Maquina
}

var fabricas = []fabrica{
	{"switch", func(a bool) Maquina { return NovoDocSwitch(a) }},
	{"State", func(a bool) Maquina { return NovoDocumento(a) }},
	{"tabela", func(a bool) Maquina { return NovoDocTabela(a) }},
}

func aplicar(m Maquina, ev Evento) error {
	if ev == EvPublicar {
		return m.Publicar()
	}
	return m.Rejeitar()
}

// Teste diferencial: todas as sequências de até 8 eventos, com e sem
// administrador, precisam dar o MESMO resultado nas três formas, e
// também na forma de funções de estado.
func TestFormasEquivalentes(t *testing.T) {
	total := 0
	for _, admin := range []bool{false, true} {
		for n := 0; n <= 8; n++ {
			for mascara := 0; mascara < 1<<n; mascara++ {
				var eventos []Evento
				for i := 0; i < n; i++ {
					eventos = append(eventos, Evento(mascara>>i&1))
				}
				var refEstado string
				var refRecusas []bool
				for i, fb := range fabricas {
					m := fb.nova(admin)
					var recusas []bool
					for _, ev := range eventos {
						err := aplicar(m, ev)
						if err != nil && !errors.Is(err, ErrTransicao) {
							t.Fatalf("%s: erro inesperado %v", fb.nome, err)
						}
						recusas = append(recusas, err != nil)
					}
					if i == 0 {
						refEstado, refRecusas = m.Estado(), recusas
						continue
					}
					if m.Estado() != refEstado {
						t.Fatalf("%s: estado %s != %s para %v (admin=%v)", fb.nome, m.Estado(), refEstado, eventos, admin)
					}
					for j := range recusas {
						if recusas[j] != refRecusas[j] {
							t.Fatalf("%s: recusa divergente no passo %d de %v", fb.nome, j, eventos)
						}
					}
				}
				f := &Fluxo{Admin: admin}
				final := f.Rodar(eventos...)
				nRec := 0
				for _, r := range refRecusas {
					if r {
						nRec++
					}
				}
				if len(eventos) > 0 && (final != refEstado || len(f.Erros) != nRec) {
					t.Fatalf("funções: final=%s recusas=%d, esperado %s/%d para %v", final, len(f.Erros), refEstado, nRec, eventos)
				}
				total++
			}
		}
	}
	t.Logf("%d sequências comparadas nas 4 formas", total)
}
