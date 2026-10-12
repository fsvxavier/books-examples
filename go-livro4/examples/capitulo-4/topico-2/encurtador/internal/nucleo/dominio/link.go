// Package dominio reúne as regras de negócio do encurtador: o que é um link,
// o que torna um endereço ou um código aceitável, e os erros que o negócio
// reconhece. Não importa nada além da biblioteca padrão e não sabe que existem
// HTTP, bancos de dados ou relógio de parede.
package dominio

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Erros do negócio. Quem chama os testa com errors.Is.
var (
	ErrDestinoInvalido = errors.New("destino inválido")
	ErrCodigoInvalido  = errors.New("código inválido")
	ErrNaoEncontrado   = errors.New("link não encontrado")
	ErrCodigoEmUso     = errors.New("código já está em uso")
)

// Regras de formato. O alfabeto exclui símbolos que se confundem na leitura.
const (
	Alfabeto          = "abcdefghjkmnpqrstuvwxyz23456789"
	TamanhoCodigo     = 6
	TamanhoMaxDestino = 2048
)

// Link é a entidade central: um endereço de destino com um código curto.
type Link struct {
	Codigo  string
	Destino string
	Criado  time.Time
	Acessos int
}

// NovoLink valida as regras e devolve um link recém-criado, sem acessos.
// O destino é devolvido na forma normalizada.
func NovoLink(codigo, destino string, criado time.Time) (Link, error) {
	if err := ValidarCodigo(codigo); err != nil {
		return Link{}, err
	}
	normalizado, err := NormalizarDestino(destino)
	if err != nil {
		return Link{}, err
	}
	return Link{Codigo: codigo, Destino: normalizado, Criado: criado}, nil
}

// ValidarCodigo confere o tamanho e o alfabeto do código.
func ValidarCodigo(codigo string) error {
	if len(codigo) != TamanhoCodigo {
		return fmt.Errorf("%w: precisa ter %d símbolos", ErrCodigoInvalido, TamanhoCodigo)
	}
	for _, r := range codigo {
		if !strings.ContainsRune(Alfabeto, r) {
			return fmt.Errorf("%w: símbolo %q fora do alfabeto", ErrCodigoInvalido, r)
		}
	}
	return nil
}

// NormalizarDestino confere o endereço e devolve a sua forma canônica.
// Só http e https são aceitos: outros esquemas (javascript:, file:) tornariam
// o encurtador uma ferramenta de ataque.
func NormalizarDestino(bruto string) (string, error) {
	bruto = strings.TrimSpace(bruto)
	if bruto == "" {
		return "", fmt.Errorf("%w: vazio", ErrDestinoInvalido)
	}
	if len(bruto) > TamanhoMaxDestino {
		return "", fmt.Errorf("%w: mais de %d bytes", ErrDestinoInvalido, TamanhoMaxDestino)
	}
	u, err := url.Parse(bruto)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDestinoInvalido, err)
	}
	switch u.Scheme {
	case "http", "https":
	case "":
		return "", fmt.Errorf("%w: faltou o esquema, use http:// ou https://", ErrDestinoInvalido)
	default:
		return "", fmt.Errorf("%w: esquema %q não é http nem https", ErrDestinoInvalido, u.Scheme)
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("%w: sem host", ErrDestinoInvalido)
	}
	return u.String(), nil
}
