// Package e06 é um validador de structs guiado por tags, escrito com reflexão.
//
// A tag é "validate", com regras separadas por vírgula:
//
//	obrigatorio   o valor não pode ser zero (texto, fatia e mapa: não pode ser vazio)
//	min=N, max=N  texto: runas; números: valor; fatia, mapa, array: tamanho
//	email         o texto, se houver, é um endereço de e-mail
//	dive          valida cada elemento de uma fatia ou array de structs
//	-             ignora o campo
//
// Structs aninhadas e ponteiros não nulos para struct são validados sozinhos.
package e06

import (
	"errors"
	"fmt"
	"net/mail"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

// ErrTagInvalida indica uma tag mal escrita ou aplicada a um tipo que não a aceita.
// É um defeito do programa, e não dos dados: aparece na primeira validação do tipo.
var ErrTagInvalida = errors.New("tag de validação inválida")

// ErrNaoEhStruct indica que o valor recebido não é uma struct nem um ponteiro para uma.
var ErrNaoEhStruct = errors.New("a validação espera uma struct ou um ponteiro para struct")

// profundidadeMaxima protege contra estruturas com ciclos de ponteiros.
const profundidadeMaxima = 32

// Falha descreve uma regra que um campo não cumpriu.
type Falha struct {
	Campo    string // caminho do campo, como "Endereco.CEP" ou "Contatos[1].Fone"
	Regra    string
	Mensagem string
}

func (f Falha) Error() string { return f.Campo + ": " + f.Mensagem }

// Falhas é o erro devolvido quando há dados inválidos: junta todas as falhas de uma vez.
type Falhas []Falha

func (fs Falhas) Error() string {
	partes := make([]string, len(fs))
	for i, f := range fs {
		partes[i] = f.Error()
	}
	return strings.Join(partes, "; ")
}

// regra verifica um campo e devolve a mensagem de falha, ou "" se está tudo certo.
type regra struct {
	nome  string
	testa func(v reflect.Value) string
}

// campo reúne o que o validador precisa saber de um campo, calculado uma vez por tipo.
type campo struct {
	indice  int
	nome    string
	regras  []regra
	dive    bool
	aninhar bool // o campo é uma struct (ou ponteiro para struct) validada pelos próprios campos
}

// plano é o resultado de ler as tags de um tipo.
type plano struct {
	campos []campo
	err    error
}

var cache sync.Map // reflect.Type -> *plano

// Validar verifica as regras da tag "validate" de v. Devolve nil se os dados são válidos,
// Falhas se há dados inválidos, ou um erro que envolve ErrTagInvalida se uma tag está errada.
func Validar(v any) error {
	return validar(v, true)
}

func validar(v any, usarCache bool) error {
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return ErrNaoEhStruct
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ErrNaoEhStruct
	}
	var falhas Falhas
	if err := validarStruct(val, "", 0, usarCache, &falhas); err != nil {
		return err
	}
	if len(falhas) > 0 {
		return falhas
	}
	return nil
}

func planoDe(t reflect.Type, usarCache bool) *plano {
	if usarCache {
		if p, ok := cache.Load(t); ok {
			return p.(*plano)
		}
	}
	p := construirPlano(t)
	if usarCache {
		cache.Store(t, p)
	}
	return p
}

func construirPlano(t reflect.Type) *plano {
	p := &plano{}
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag, tem := f.Tag.Lookup("validate")
		if tag == "-" {
			continue
		}
		c := campo{indice: i, nome: f.Name}
		ft := f.Type
		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		c.aninhar = ft.Kind() == reflect.Struct
		if tem {
			for _, parte := range strings.Split(tag, ",") {
				nome, arg, _ := strings.Cut(strings.TrimSpace(parte), "=")
				if nome == "dive" {
					c.dive = true
					continue
				}
				r, err := novaRegra(nome, arg, f.Type)
				if err != nil {
					p.err = fmt.Errorf("%w: %s.%s: %v", ErrTagInvalida, t.Name(), f.Name, err)
					return p
				}
				c.regras = append(c.regras, r)
			}
		}
		if len(c.regras) > 0 || c.dive || c.aninhar {
			p.campos = append(p.campos, c)
		}
	}
	return p
}

func novaRegra(nome, arg string, t reflect.Type) (regra, error) {
	switch nome {
	case "obrigatorio":
		return regra{nome, func(v reflect.Value) string {
			if v.IsZero() || (v.Kind() == reflect.String || v.Kind() == reflect.Slice || v.Kind() == reflect.Map) && v.Len() == 0 {
				return "é obrigatório"
			}
			return ""
		}}, nil
	case "min", "max":
		n, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return regra{}, fmt.Errorf("%s=%q: o argumento deve ser um número", nome, arg)
		}
		medida, err := medidor(t)
		if err != nil {
			return regra{}, fmt.Errorf("%s não se aplica a %s", nome, t)
		}
		if nome == "min" {
			return regra{nome, func(v reflect.Value) string {
				if m := medida(v); m < n {
					return fmt.Sprintf("deve ser no mínimo %v (é %v)", n, m)
				}
				return ""
			}}, nil
		}
		return regra{nome, func(v reflect.Value) string {
			if m := medida(v); m > n {
				return fmt.Sprintf("deve ser no máximo %v (é %v)", n, m)
			}
			return ""
		}}, nil
	case "email":
		if t.Kind() != reflect.String {
			return regra{}, fmt.Errorf("email não se aplica a %s", t)
		}
		return regra{nome, func(v reflect.Value) string {
			s := v.String()
			if s == "" {
				return "" // quem exige o campo usa "obrigatorio"
			}
			if a, err := mail.ParseAddress(s); err != nil || a.Address != s {
				return "não é um e-mail válido"
			}
			return ""
		}}, nil
	}
	return regra{}, fmt.Errorf("regra desconhecida %q", nome)
}

// medidor devolve a função que mede um valor do tipo t, ou um erro se o tipo não tem medida.
func medidor(t reflect.Type) (func(reflect.Value) float64, error) {
	switch t.Kind() {
	case reflect.String:
		return func(v reflect.Value) float64 { return float64(utf8.RuneCountInString(v.String())) }, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(v reflect.Value) float64 { return float64(v.Int()) }, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(v reflect.Value) float64 { return float64(v.Uint()) }, nil
	case reflect.Float32, reflect.Float64:
		return func(v reflect.Value) float64 { return v.Float() }, nil
	case reflect.Slice, reflect.Map, reflect.Array:
		return func(v reflect.Value) float64 { return float64(v.Len()) }, nil
	}
	return nil, errors.New("tipo sem medida")
}

func validarStruct(val reflect.Value, caminho string, profundidade int, usarCache bool, falhas *Falhas) error {
	if profundidade > profundidadeMaxima {
		return fmt.Errorf("estrutura com mais de %d níveis de profundidade (ciclo de ponteiros?)", profundidadeMaxima)
	}
	p := planoDe(val.Type(), usarCache)
	if p.err != nil {
		return p.err
	}
	for _, c := range p.campos {
		fv := val.Field(c.indice)
		nome := c.nome
		if caminho != "" {
			nome = caminho + "." + c.nome
		}
		for _, r := range c.regras {
			if msg := r.testa(fv); msg != "" {
				*falhas = append(*falhas, Falha{Campo: nome, Regra: r.nome, Mensagem: msg})
			}
		}
		if c.aninhar {
			if err := aninhado(fv, nome, profundidade, usarCache, falhas); err != nil {
				return err
			}
		}
		if c.dive && (fv.Kind() == reflect.Slice || fv.Kind() == reflect.Array) {
			for i := range fv.Len() {
				if err := aninhado(fv.Index(i), fmt.Sprintf("%s[%d]", nome, i), profundidade, usarCache, falhas); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// aninhado valida v se for uma struct ou um ponteiro não nulo para struct.
func aninhado(v reflect.Value, caminho string, profundidade int, usarCache bool, falhas *Falhas) error {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	return validarStruct(v, caminho, profundidade+1, usarCache, falhas)
}
