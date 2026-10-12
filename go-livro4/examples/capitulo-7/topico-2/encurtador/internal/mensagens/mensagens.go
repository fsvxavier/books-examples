// Package mensagens define a LINGUAGEM PUBLICADA do encurtador: o formato das
// mensagens que cruzam a fronteira do processo. Os eventos de domínio
// (dominio.LinkCriado e outros) são do modelo e podem mudar com ele; as
// mensagens são um contrato com quem as consome, com formato estável e versão.
package mensagens

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-4/capitulo-7/topico-2/encurtador/internal/dominio"
)

// Os tópicos têm o tipo e a versão no nome: uma mudança incompatível vira um
// tópico novo (.v2), e quem consome a versão antiga continua funcionando.
const (
	TopicoLinkCriado   = "encurtador.link_criado.v1"
	TopicoLinkAcessado = "encurtador.link_acessado.v1"
	TopicoLinkExpirado = "encurtador.link_expirado.v1"
)

// Envelope é o que viaja: um identificador único da mensagem (para quem
// consome desfazer duplicatas), o instante do fato e os dados em JSON.
type Envelope struct {
	ID         string          `json:"id"`
	Tipo       string          `json:"tipo"`
	OcorridoEm time.Time       `json:"ocorrido_em"`
	Dados      json.RawMessage `json:"dados"`
}

// LinkCriadoV1 é o conteúdo de TopicoLinkCriado.
type LinkCriadoV1 struct {
	Codigo  string     `json:"codigo"`
	Destino string     `json:"destino"`
	Expira  *time.Time `json:"expira,omitempty"`
}

// LinkAcessadoV1 é o conteúdo de TopicoLinkAcessado.
type LinkAcessadoV1 struct {
	Codigo string `json:"codigo"`
}

// LinkExpiradoV1 é o conteúdo de TopicoLinkExpirado.
type LinkExpiradoV1 struct {
	Codigo string `json:"codigo"`
}

// Novo monta um envelope com os dados codificados.
func Novo(id, tipo string, ocorridoEm time.Time, dados any) (Envelope, error) {
	bytes, err := json.Marshal(dados)
	if err != nil {
		return Envelope{}, fmt.Errorf("codificando %s: %w", tipo, err)
	}
	return Envelope{ID: id, Tipo: tipo, OcorridoEm: ocorridoEm, Dados: bytes}, nil
}

// DeEvento traduz um evento de domínio para a mensagem publicada. É o ÚNICO
// ponto de tradução entre o modelo interno e o contrato externo: se o domínio
// mudar, muda aqui, e quem consome não percebe.
func DeEvento(id string, e dominio.Evento) (topico string, env Envelope, err error) {
	switch e := e.(type) {
	case dominio.LinkCriado:
		dados := LinkCriadoV1{Codigo: e.Codigo.String(), Destino: e.Destino.String()}
		if !e.Expira.IsZero() {
			dados.Expira = &e.Expira
		}
		env, err = Novo(id, "link_criado", e.Em, dados)
		return TopicoLinkCriado, env, err
	case dominio.LinkAcessado:
		env, err = Novo(id, "link_acessado", e.Em, LinkAcessadoV1{Codigo: e.Codigo.String()})
		return TopicoLinkAcessado, env, err
	default:
		return "", Envelope{}, fmt.Errorf("evento %q não tem mensagem publicada", e.Nome())
	}
}

// Decodificar lê os dados do envelope no tipo pedido.
func Decodificar[T any](env Envelope) (T, error) {
	var v T
	if err := json.Unmarshal(env.Dados, &v); err != nil {
		return v, fmt.Errorf("decodificando %s: %w", env.Tipo, err)
	}
	return v, nil
}

// Bytes codifica o envelope para trafegar.
func (e Envelope) Bytes() ([]byte, error) { return json.Marshal(e) }

// DoBytes lê um envelope que veio da rede.
func DoBytes(b []byte) (Envelope, error) {
	var e Envelope
	if err := json.Unmarshal(b, &e); err != nil {
		return Envelope{}, fmt.Errorf("lendo o envelope: %w", err)
	}
	if e.ID == "" || e.Tipo == "" {
		return Envelope{}, fmt.Errorf("envelope sem id ou tipo: %s", b)
	}
	return e, nil
}

// NovoID devolve um identificador aleatório de 128 bits em hexadecimal, no
// formato de um UUID versão 4, para identificar uma mensagem. Quem consome o
// usa para reconhecer entregas repetidas.
func NovoID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40 // versão 4
	b[8] = b[8]&0x3f | 0x80 // variante RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
