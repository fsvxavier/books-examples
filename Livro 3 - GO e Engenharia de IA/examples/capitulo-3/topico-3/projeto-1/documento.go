// Package registro define o documento de entrada e a validação do seu esquema.
package registro

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Documento é o contrato de entrada do pipeline.
type Documento struct {
	ID     string `json:"id"`
	Fonte  string `json:"fonte"`
	Titulo string `json:"titulo"`
	Texto  string `json:"texto"`
	Idioma string `json:"idioma"`
}

// Limites do esquema.
const (
	MaxTexto  = 2 << 20 // 2 MiB
	MaxTitulo = 300
)

// ErroCampo diz qual campo falhou e por quê.
type ErroCampo struct {
	Campo  string
	Motivo string
}

func (e ErroCampo) Error() string { return e.Campo + ": " + e.Motivo }

// ErroEsquema indica que o registro é sintaticamente JSON mas não obedece ao
// contrato. Não adianta tentar de novo: o dado precisa ser corrigido.
var ErroEsquema = errors.New("registro fora do esquema")

// Decodificar lê um registro JSON, rejeita campos desconhecidos e valida todos
// os campos de uma vez, devolvendo todos os problemas juntos.
func Decodificar(bruto []byte) (Documento, error) {
	var d Documento
	dec := json.NewDecoder(bytes.NewReader(bruto))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return Documento{}, fmt.Errorf("%w: json: %w", ErroEsquema, err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return Documento{}, fmt.Errorf("%w: json: lixo depois do objeto", ErroEsquema)
	}
	if err := d.Validar(); err != nil {
		return Documento{}, err
	}
	return d, nil
}

// Validar confere o contrato e junta todos os erros encontrados.
func (d Documento) Validar() error {
	var erros []error
	falha := func(campo, motivo string) {
		erros = append(erros, ErroCampo{campo, motivo})
	}
	if strings.TrimSpace(d.ID) == "" {
		falha("id", "obrigatório")
	}
	if strings.TrimSpace(d.Fonte) == "" {
		falha("fonte", "obrigatório")
	}
	if n := utf8.RuneCountInString(d.Titulo); n == 0 || n > MaxTitulo {
		falha("titulo", fmt.Sprintf("tamanho %d fora de 1..%d", n, MaxTitulo))
	}
	switch {
	case strings.TrimSpace(d.Texto) == "":
		falha("texto", "vazio")
	case len(d.Texto) > MaxTexto:
		falha("texto", fmt.Sprintf("%d bytes passa de %d", len(d.Texto), MaxTexto))
	case !utf8.ValidString(d.Texto):
		falha("texto", "UTF-8 inválido")
	}
	switch d.Idioma {
	case "pt", "en", "es":
	default:
		falha("idioma", fmt.Sprintf("%q não suportado", d.Idioma))
	}
	if len(erros) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %w", ErroEsquema, errors.Join(erros...))
}
