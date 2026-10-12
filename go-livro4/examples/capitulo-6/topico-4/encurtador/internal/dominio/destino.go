package dominio

import (
	"fmt"
	"net/url"
	"strings"
)

// TamanhoMaxDestino limita o endereço, para proteger quem o guarda.
const TamanhoMaxDestino = 2048

// Destino é um objeto de valor: o endereço para onde o link aponta, já
// normalizado. Só http e https são aceitos: outros esquemas (javascript:,
// file:) tornariam o encurtador uma ferramenta de ataque.
type Destino struct {
	endereco string
}

// NovoDestino confere o endereço e guarda a sua forma canônica.
func NovoDestino(bruto string) (Destino, error) {
	bruto = strings.TrimSpace(bruto)
	if bruto == "" {
		return Destino{}, fmt.Errorf("%w: vazio", ErrDestinoInvalido)
	}
	if len(bruto) > TamanhoMaxDestino {
		return Destino{}, fmt.Errorf("%w: mais de %d bytes", ErrDestinoInvalido, TamanhoMaxDestino)
	}
	u, err := url.Parse(bruto)
	if err != nil {
		return Destino{}, fmt.Errorf("%w: %v", ErrDestinoInvalido, err)
	}
	switch u.Scheme {
	case "http", "https":
	case "":
		return Destino{}, fmt.Errorf("%w: faltou o esquema, use http:// ou https://", ErrDestinoInvalido)
	default:
		return Destino{}, fmt.Errorf("%w: esquema %q não é http nem https", ErrDestinoInvalido, u.Scheme)
	}
	if u.Hostname() == "" {
		return Destino{}, fmt.Errorf("%w: sem host", ErrDestinoInvalido)
	}
	return Destino{endereco: u.String()}, nil
}

// String devolve o endereço normalizado.
func (d Destino) String() string { return d.endereco }

// EhZero diz se é o valor zero, que não é um destino válido.
func (d Destino) EhZero() bool { return d.endereco == "" }
