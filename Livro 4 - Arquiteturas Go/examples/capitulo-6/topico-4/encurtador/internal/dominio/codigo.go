package dominio

import (
	"fmt"
	"strings"
)

// Regras de formato do código. O alfabeto exclui símbolos que se confundem na leitura.
const (
	Alfabeto      = "abcdefghjkmnpqrstuvwxyz23456789"
	TamanhoCodigo = 6
)

// Codigo é um objeto de valor: o código curto de um link. O campo é privado,
// então a única forma de obter um Codigo que não seja o valor zero é NovoCodigo,
// que confere as regras. Um Codigo em circulação é, por construção, válido.
type Codigo struct {
	valor string
}

// NovoCodigo confere o tamanho e o alfabeto.
func NovoCodigo(texto string) (Codigo, error) {
	if len(texto) != TamanhoCodigo {
		return Codigo{}, fmt.Errorf("%w: precisa ter %d símbolos", ErrCodigoInvalido, TamanhoCodigo)
	}
	for _, r := range texto {
		if !strings.ContainsRune(Alfabeto, r) {
			return Codigo{}, fmt.Errorf("%w: símbolo %q fora do alfabeto", ErrCodigoInvalido, r)
		}
	}
	return Codigo{valor: texto}, nil
}

// String devolve o texto do código.
func (c Codigo) String() string { return c.valor }

// EhZero diz se é o valor zero, que não é um código válido.
func (c Codigo) EhZero() bool { return c.valor == "" }
