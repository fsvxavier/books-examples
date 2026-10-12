// Package e12 reescreve o Decodificar do encurtador (Capítulo 46) em três versões.
package e12

import (
	"encoding/json"
	"fmt"
	"time"
)

// Envelope é o que viaja entre os serviços.
type Envelope struct {
	ID         string          `json:"id"`
	Tipo       string          `json:"tipo"`
	OcorridoEm time.Time       `json:"ocorrido_em"`
	Dados      json.RawMessage `json:"dados"`
}

// LinkCriadoV1 e LinkAcessadoV1 são dois conteúdos possíveis.
type LinkCriadoV1 struct {
	Codigo  string `json:"codigo"`
	Destino string `json:"destino"`
}

// LinkAcessadoV1 é o conteúdo de um acesso.
type LinkAcessadoV1 struct {
	Codigo string `json:"codigo"`
}

// Tipo dá o nome da mensagem, como aparece no envelope.
func (LinkCriadoV1) Tipo() string { return "link_criado" }

// Tipo dá o nome da mensagem, como aparece no envelope.
func (LinkAcessadoV1) Tipo() string { return "link_acessado" }

// DecodificarInto é a versão sem generics: o chamador passa um ponteiro e o
// compilador não confere o que ele passa.
func DecodificarInto(env Envelope, destino any) error {
	if err := json.Unmarshal(env.Dados, destino); err != nil {
		return fmt.Errorf("decodificando %s: %w", env.Tipo, err)
	}
	return nil
}

// Decodificar é a versão do Capítulo 46: o tipo T é o resultado.
func Decodificar[T any](env Envelope) (T, error) {
	var v T
	if err := json.Unmarshal(env.Dados, &v); err != nil {
		return v, fmt.Errorf("decodificando %s: %w", env.Tipo, err)
	}
	return v, nil
}

// Mensagem é a constraint da versão segura: T sabe dizer o próprio tipo.
type Mensagem interface{ Tipo() string }

// DecodificarSeguro confere o tipo do envelope antes de decodificar.
func DecodificarSeguro[T Mensagem](env Envelope) (T, error) {
	var v T
	if env.Tipo != v.Tipo() {
		return v, fmt.Errorf("o envelope é %q, e %T espera %q", env.Tipo, v, v.Tipo())
	}
	if err := json.Unmarshal(env.Dados, &v); err != nil {
		return v, fmt.Errorf("decodificando %s: %w", env.Tipo, err)
	}
	return v, nil
}
