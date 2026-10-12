// Package cobranca cobra clientes usando os meios de pagamento registrados.
package cobranca

import (
	"errors"
	"fmt"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-14/projeto-1"
	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-16/projeto-1/config"
)

// Meio e a capacidade de cobrar um valor.
type Meio interface {
	Cobrar(centavos int64) (comprovante string, err error)
}

// Catalogo e o registro dos meios, ja preenchido pela raiz de composicao.
type Catalogo = registro.Registro[Meio]

var (
	ErrMeioDesconhecido = errors.New("meio de pagamento desconhecido")
	ErrAcimaDoLimite    = errors.New("valor acima do limite")
)

// Servico recebe tudo de que precisa pelo construtor.
type Servico struct {
	cfg      config.Config
	catalogo *Catalogo
}

func NovoServico(cfg config.Config, catalogo *Catalogo) *Servico {
	return &Servico{cfg: cfg, catalogo: catalogo}
}

func (s *Servico) Cobrar(nomeMeio string, centavos int64) (string, error) {
	if centavos > s.cfg.LimiteCentavos {
		return "", fmt.Errorf("%w: %d > %d", ErrAcimaDoLimite, centavos, s.cfg.LimiteCentavos)
	}
	m, ok := s.catalogo.Obter(nomeMeio)
	if !ok {
		return "", fmt.Errorf("%w: %q (disponiveis: %v)", ErrMeioDesconhecido, nomeMeio, s.catalogo.Nomes())
	}
	return m.Cobrar(centavos)
}
