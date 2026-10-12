package lsp

import "errors"

// ---- Violacao classica: o quadrado "e um" retangulo? ----

type Retangulo interface {
	DefinirLargura(w int)
	DefinirAltura(h int)
	Area() int
}

type Ret struct{ w, h int }

func (r *Ret) DefinirLargura(w int) { r.w = w }
func (r *Ret) DefinirAltura(h int)  { r.h = h }
func (r *Ret) Area() int            { return r.w * r.h }

// Quadrado satisfaz a interface pelos tipos, mas QUEBRA o contrato:
// quem muda a largura espera que a altura fique como estava.
type Quadrado struct{ lado int }

func (q *Quadrado) DefinirLargura(w int) { q.lado = w }
func (q *Quadrado) DefinirAltura(h int)  { q.lado = h }
func (q *Quadrado) Area() int            { return q.lado * q.lado }

// ---- Outra violacao: o armazenamento "somente leitura" ----

var ErrSomenteLeitura = errors.New("armazenamento somente leitura")

type Armazem interface {
	// Guardar persiste o valor; depois de um Guardar sem erro, Ler(chave) devolve o valor.
	Guardar(chave, valor string) error
	Ler(chave string) (string, bool)
}

type Memoria struct{ m map[string]string }

func NovaMemoria() *Memoria { return &Memoria{m: map[string]string{}} }

func (a *Memoria) Guardar(k, v string) error { a.m[k] = v; return nil }
func (a *Memoria) Ler(k string) (string, bool) {
	v, ok := a.m[k]
	return v, ok
}

// Ruim: finge ter sucesso e nao guarda nada. Mentir e pior do que recusar.
type SomenteLeituraMentirosa struct{ Memoria }

func (SomenteLeituraMentirosa) Guardar(k, v string) error { return nil }

// Melhor: recusa com um erro claro, que o contrato documenta.
type SomenteLeitura struct{ *Memoria }

func (SomenteLeitura) Guardar(string, string) error { return ErrSomenteLeitura }

// Melhor ainda: nao tem Guardar. O tipo nem promete o que nao cumpre.
type Leitor interface {
	Ler(chave string) (string, bool)
}
