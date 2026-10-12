package e04

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrDivisorZero = errors.New("divisor zero")

func DividirPanic(a, b int) int {
	if b == 0 {
		panic("divisão por zero")
	}
	return a / b
}

// Seguro traduz o pânico de volta em erro: o chamador paga o custo de recover só para tratar caso normal.
func DividirPanicSeguro(a, b int) (r int, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("recuperado: %v", p)
		}
	}()
	return DividirPanic(a, b), nil
}

func DividirErro(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivisorZero
	}
	return a / b, nil
}

// ---- Uso legítimo: pânico interno, erro na fronteira pública (como encoding/json e text/template). ----

type erroSintaxe struct {
	pos int
	msg string
}

func (e erroSintaxe) Error() string { return "posição " + strconv.Itoa(e.pos) + ": " + e.msg }

type analisador struct {
	entrada []byte
	pos     int
}

func (a *analisador) falhar(msg string) {
	panic(erroSintaxe{a.pos, msg}) // atravessa as chamadas recursivas sem if err != nil em cada nível
}

func (a *analisador) lista() []int {
	var v []int
	a.esperar('[')
	for {
		v = append(v, a.numero())
		if a.entrada[a.pos] == ']' {
			a.pos++
			return v
		}
		a.esperar(',')
	}
}

func (a *analisador) esperar(c byte) {
	if a.pos >= len(a.entrada) || a.entrada[a.pos] != c {
		a.falhar("esperava '" + string(c) + "'")
	}
	a.pos++
}

func (a *analisador) numero() int {
	ini := a.pos
	for a.pos < len(a.entrada) && a.entrada[a.pos] >= '0' && a.entrada[a.pos] <= '9' {
		a.pos++
	}
	if ini == a.pos {
		a.falhar("esperava dígito")
	}
	n, _ := strconv.Atoi(string(a.entrada[ini:a.pos]))
	return n
}

// Analisar é a API pública: nenhum pânico escapa daqui.
func Analisar(s string) (v []int, err error) {
	defer func() {
		if p := recover(); p != nil {
			if e, ok := p.(erroSintaxe); ok {
				err = e
				return
			}
			panic(p) // bug de verdade (índice fora dos limites etc.): não esconder
		}
	}()
	a := &analisador{entrada: []byte(s)}
	return a.lista(), nil
}
