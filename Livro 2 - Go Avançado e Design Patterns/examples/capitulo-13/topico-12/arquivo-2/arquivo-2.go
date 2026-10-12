package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Expr é a árvore: uma interface selada (só este pacote cria nós).
type Expr interface{ isExpr() }

type (
	Num  struct{ V float64 }
	Var  struct{ Nome string }
	Soma struct{ E, D Expr }
	Mul  struct{ E, D Expr }
	Neg  struct{ E Expr }
)

func (Num) isExpr()  {}
func (Var) isExpr()  {}
func (Soma) isExpr() {}
func (Mul) isExpr()  {}
func (Neg) isExpr()  {}

// VisitanteR devolve um R: cada operação escolhe o seu tipo de resultado.
type VisitanteR[R any] interface {
	VisitarNum(Num) R
	VisitarVar(Var) R
	VisitarSoma(Soma) R
	VisitarMul(Mul) R
	VisitarNeg(Neg) R
}

// Aceitar faz o despacho. Os filhos são visitados PELO VISITANTE (ele
// chama Aceitar de volta), que decide se desce e como combina.
func Aceitar[R any](e Expr, v VisitanteR[R]) R {
	switch e := e.(type) {
	case Num:
		return v.VisitarNum(e)
	case Var:
		return v.VisitarVar(e)
	case Soma:
		return v.VisitarSoma(e)
	case Mul:
		return v.VisitarMul(e)
	case Neg:
		return v.VisitarNeg(e)
	}
	panic(fmt.Sprintf("Aceitar: tipo %T não tratado", e))
}

// ---- Operação 1: imprimir (R = string).
type Imprimir struct{}

func (Imprimir) VisitarNum(n Num) string { return fmt.Sprint(n.V) }
func (Imprimir) VisitarVar(v Var) string { return v.Nome }
func (i Imprimir) VisitarSoma(s Soma) string {
	return "(" + Aceitar[string](s.E, i) + " + " + Aceitar[string](s.D, i) + ")"
}
func (i Imprimir) VisitarMul(m Mul) string {
	return "(" + Aceitar[string](m.E, i) + " * " + Aceitar[string](m.D, i) + ")"
}
func (i Imprimir) VisitarNeg(n Neg) string { return "-" + Aceitar[string](n.E, i) }

// ---- Operação 2: avaliar (R = Resultado, com erro).
type Resultado struct {
	V   float64
	Err error
}

var ErrVariavel = errors.New("variável sem valor")

type Avaliar struct{ Env map[string]float64 }

func (Avaliar) VisitarNum(n Num) Resultado { return Resultado{V: n.V} }
func (a Avaliar) VisitarVar(v Var) Resultado {
	x, ok := a.Env[v.Nome]
	if !ok {
		return Resultado{Err: fmt.Errorf("%w: %s", ErrVariavel, v.Nome)}
	}
	return Resultado{V: x}
}
func (a Avaliar) bin(e, d Expr, op func(x, y float64) float64) Resultado {
	l := Aceitar[Resultado](e, a)
	if l.Err != nil {
		return l
	}
	r := Aceitar[Resultado](d, a)
	if r.Err != nil {
		return r
	}
	return Resultado{V: op(l.V, r.V)}
}
func (a Avaliar) VisitarSoma(s Soma) Resultado {
	return a.bin(s.E, s.D, func(x, y float64) float64 { return x + y })
}
func (a Avaliar) VisitarMul(m Mul) Resultado {
	return a.bin(m.E, m.D, func(x, y float64) float64 { return x * y })
}
func (a Avaliar) VisitarNeg(n Neg) Resultado {
	r := Aceitar[Resultado](n.E, a)
	r.V = -r.V
	return r
}

// ---- Operação 3: simplificar (R = Expr: devolve OUTRA árvore).
type Simplificar struct{}

func (Simplificar) VisitarNum(n Num) Expr { return n }
func (Simplificar) VisitarVar(v Var) Expr { return v }
func (s Simplificar) VisitarSoma(x Soma) Expr {
	e, d := Aceitar[Expr](x.E, s), Aceitar[Expr](x.D, s)
	en, eNum := e.(Num)
	dn, dNum := d.(Num)
	switch {
	case eNum && dNum:
		return Num{en.V + dn.V} // dobra constantes
	case eNum && en.V == 0:
		return d
	case dNum && dn.V == 0:
		return e
	}
	return Soma{e, d}
}
func (s Simplificar) VisitarMul(x Mul) Expr {
	e, d := Aceitar[Expr](x.E, s), Aceitar[Expr](x.D, s)
	en, eNum := e.(Num)
	dn, dNum := d.(Num)
	switch {
	case eNum && dNum:
		return Num{en.V * dn.V}
	case (eNum && en.V == 0) || (dNum && dn.V == 0):
		return Num{0}
	case eNum && en.V == 1:
		return d
	case dNum && dn.V == 1:
		return e
	}
	return Mul{e, d}
}
func (s Simplificar) VisitarNeg(n Neg) Expr {
	e := Aceitar[Expr](n.E, s)
	if num, ok := e.(Num); ok {
		return Num{-num.V}
	}
	return Neg{e}
}

// ---- Operação 4: coletar variáveis (R = struct{}; acumula no campo).
type Variaveis struct{ Vistas map[string]bool }

func (Variaveis) VisitarNum(Num) struct{} { return struct{}{} }
func (v Variaveis) VisitarVar(x Var) struct{} {
	v.Vistas[x.Nome] = true
	return struct{}{}
}
func (v Variaveis) VisitarSoma(s Soma) struct{} {
	Aceitar[struct{}](s.E, v)
	return Aceitar[struct{}](s.D, v)
}
func (v Variaveis) VisitarMul(m Mul) struct{} {
	Aceitar[struct{}](m.E, v)
	return Aceitar[struct{}](m.D, v)
}
func (v Variaveis) VisitarNeg(n Neg) struct{} { return Aceitar[struct{}](n.E, v) }

func main() {
	// (x + 0) * 1 + 2 * 3 + (y * 0)
	e := Soma{
		Soma{Mul{Soma{Var{"x"}, Num{0}}, Num{1}}, Mul{Num{2}, Num{3}}},
		Mul{Var{"y"}, Num{0}},
	}
	env := map[string]float64{"x": 5, "y": 7}

	fmt.Println("original:   ", Aceitar[string](e, Imprimir{}))
	s := Aceitar[Expr](e, Simplificar{})
	fmt.Println("simplificada:", Aceitar[string](s, Imprimir{}))

	r1 := Aceitar[Resultado](e, Avaliar{env})
	r2 := Aceitar[Resultado](s, Avaliar{env})
	fmt.Println("avaliação x=5, y=7:", r1.V, "|", r2.V)
	fmt.Println("variável ausente:   ", Aceitar[Resultado](e, Avaliar{map[string]float64{"x": 1}}).Err)

	vs := Variaveis{map[string]bool{}}
	Aceitar[struct{}](e, vs)
	var nomes []string
	for n := range vs.Vistas {
		nomes = append(nomes, n)
	}
	sort.Strings(nomes)
	fmt.Println("variáveis:", strings.Join(nomes, ", "))
}
