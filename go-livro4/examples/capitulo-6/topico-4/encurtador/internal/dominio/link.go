// Package dominio reúne o modelo do negócio do encurtador: objetos de valor
// (Codigo, Destino), o agregado Link, os eventos e os erros. Não importa nada
// além da biblioteca padrão.
package dominio

import (
	"fmt"
	"time"
)

// Link é a raiz do agregado: um endereço de destino com um código curto e,
// opcionalmente, um prazo de validade. Todos os campos são privados: o estado
// só muda por métodos que respeitam as regras, e um Link só nasce por NovoLink
// ou Reconstituir.
//
// O contador de acessos NÃO faz parte do agregado. Ele não protege nenhuma
// regra do Link, muda a cada visita e, dentro de um agregado com controle de
// versão, derrubaria a concorrência (veja o capítulo). Quem conta acessos
// reage ao evento LinkAcessado.
type Link struct {
	codigo  Codigo
	destino Destino
	criado  time.Time
	expira  time.Time // zero significa "sem prazo"
	eventos []Evento
}

// NovoLink cria um link. validade zero significa que ele não expira; negativa
// é recusada. O link já nasce com o evento LinkCriado pendente.
func NovoLink(codigo Codigo, destino Destino, criado time.Time, validade time.Duration) (Link, error) {
	if codigo.EhZero() {
		return Link{}, fmt.Errorf("%w: código ausente", ErrCodigoInvalido)
	}
	if destino.EhZero() {
		return Link{}, fmt.Errorf("%w: destino ausente", ErrDestinoInvalido)
	}
	if validade < 0 {
		return Link{}, fmt.Errorf("%w: %v é negativa", ErrValidadeInvalida, validade)
	}
	l := Link{codigo: codigo, destino: destino, criado: criado}
	if validade > 0 {
		l.expira = criado.Add(validade)
	}
	l.eventos = append(l.eventos, LinkCriado{Codigo: codigo, Destino: destino, Em: criado})
	return l, nil
}

// Reconstituir recria um link a partir do que o armazenamento guardou. Reaplica
// as regras dos objetos de valor, mas não gera eventos: nada de novo aconteceu.
// expira zero significa "sem prazo".
func Reconstituir(codigo, destino string, criado, expira time.Time) (Link, error) {
	c, err := NovoCodigo(codigo)
	if err != nil {
		return Link{}, err
	}
	d, err := NovoDestino(destino)
	if err != nil {
		return Link{}, err
	}
	return Link{codigo: c, destino: d, criado: criado, expira: expira}, nil
}

// Codigo devolve o código curto.
func (l Link) Codigo() Codigo { return l.codigo }

// Destino devolve o endereço de destino.
func (l Link) Destino() Destino { return l.destino }

// Criado devolve o instante da criação.
func (l Link) Criado() time.Time { return l.criado }

// Expira devolve o instante em que o link deixa de valer. O segundo valor é
// falso para um link sem prazo.
func (l Link) Expira() (time.Time, bool) { return l.expira, !l.expira.IsZero() }

// Expirado diz se, no instante dado, o prazo já passou.
func (l Link) Expirado(agora time.Time) bool {
	return !l.expira.IsZero() && !agora.Before(l.expira)
}

// Acessar aplica a regra "um link expirado não é seguido". Se o link está
// valendo, registra o evento LinkAcessado; se não, devolve ErrExpirado e não
// registra nada.
func (l *Link) Acessar(agora time.Time) error {
	if l.Expirado(agora) {
		return ErrExpirado
	}
	l.eventos = append(l.eventos, LinkAcessado{Codigo: l.codigo, Em: agora})
	return nil
}

// RecolherEventos devolve os eventos pendentes e os esquece: cada evento é
// entregue uma vez só.
func (l *Link) RecolherEventos() []Evento {
	e := l.eventos
	l.eventos = nil
	return e
}
