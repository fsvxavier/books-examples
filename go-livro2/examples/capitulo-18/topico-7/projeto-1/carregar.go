// Package e09 preenche uma struct de configuração a partir de variáveis de ambiente, com reflexão.
//
// Tags de cada campo:
//
//	env:"NOME"       a variável de ambiente (em uma struct aninhada, é o prefixo das variáveis dela)
//	env:"NOME,obrigatorio"   falha se a variável não existir e não houver padrão
//	padrao:"valor"   o valor usado quando a variável não existe
//
// Tipos aceitos: string, bool, inteiros, floats, time.Duration, []string (separado por vírgulas),
// qualquer tipo que implemente encoding.TextUnmarshaler e structs aninhadas.
package e09

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	tipoDuracao    = reflect.TypeFor[time.Duration]()
	tipoTextoUnm   = reflect.TypeFor[encoding.TextUnmarshaler]()
	errDestino     = errors.New("o destino deve ser um ponteiro não nulo para uma struct")
	ErrObrigatoria = errors.New("variável obrigatória ausente")
)

// Carregar preenche destino lendo as variáveis com getenv. Todos os problemas são
// devolvidos juntos (errors.Join), para o operador corrigir a configuração de uma vez.
func Carregar(destino any, getenv func(string) (string, bool)) error {
	v := reflect.ValueOf(destino)
	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return errDestino
	}
	var erros []error
	preencher(v.Elem(), "", getenv, &erros)
	return errors.Join(erros...)
}

func preencher(v reflect.Value, prefixo string, getenv func(string) (string, bool), erros *[]error) {
	for f, fv := range v.Fields() {
		if !f.IsExported() {
			continue
		}
		tag, tem := f.Tag.Lookup("env")
		if !tem {
			continue
		}
		nome, opcoes, _ := strings.Cut(tag, ",")
		if f.Type.Kind() == reflect.Struct && !implementaTexto(f.Type) && f.Type != tipoDuracao {
			preencher(fv, prefixo+nome, getenv, erros)
			continue
		}
		chave := prefixo + nome
		texto, achou := getenv(chave)
		if !achou {
			padrao, temPadrao := f.Tag.Lookup("padrao")
			switch {
			case temPadrao:
				texto = padrao
			case opcoes == "obrigatorio":
				*erros = append(*erros, fmt.Errorf("%w: %s", ErrObrigatoria, chave))
				continue
			default:
				continue // fica com o valor zero
			}
		}
		if err := converter(fv, texto); err != nil {
			*erros = append(*erros, fmt.Errorf("%s=%q (campo %s): %w", chave, texto, f.Name, err))
		}
	}
}

func implementaTexto(t reflect.Type) bool {
	return reflect.PointerTo(t).Implements(tipoTextoUnm)
}

// converter grava o texto no campo, conforme o tipo do campo.
func converter(fv reflect.Value, texto string) error {
	if implementaTexto(fv.Type()) {
		return fv.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(texto))
	}
	switch {
	case fv.Type() == tipoDuracao:
		d, err := time.ParseDuration(texto)
		if err != nil {
			return err
		}
		fv.SetInt(int64(d))
		return nil
	}
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(texto)
	case reflect.Bool:
		b, err := strconv.ParseBool(texto)
		if err != nil {
			return err
		}
		fv.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(texto, 10, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(texto, 10, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(texto, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetFloat(n)
	case reflect.Slice:
		if fv.Type().Elem().Kind() != reflect.String {
			return fmt.Errorf("fatia de %s não é suportada", fv.Type().Elem())
		}
		partes := strings.Split(texto, ",")
		fatia := reflect.MakeSlice(fv.Type(), len(partes), len(partes))
		for i, p := range partes {
			fatia.Index(i).SetString(strings.TrimSpace(p))
		}
		fv.Set(fatia)
	default:
		return fmt.Errorf("tipo %s não suportado", fv.Type())
	}
	return nil
}
