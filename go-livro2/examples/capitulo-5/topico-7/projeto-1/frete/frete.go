// Package frete define o contrato de cotacao e adapta os SDKs legados a ele.
package frete

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-5/topico-7/projeto-1/legado"
)

type Pedido struct {
	CEPOrigem, CEPDestino string
	PesoGramas            int
}

type Oferta struct {
	Transportadora string
	Centavos       int64
	Prazo          time.Duration
}

// Transportadora e a interface do cliente: centavos, time.Duration, context.
type Transportadora interface {
	Nome() string
	Cotar(ctx context.Context, p Pedido) (Oferta, error)
}

// ---------- Adaptador A: texto com virgula -> centavos ----------

type AdaptadorCorreios struct{ sdk legado.CorreiosLike }

func (AdaptadorCorreios) Nome() string { return "correios-like" }

func (a AdaptadorCorreios) Cotar(ctx context.Context, p Pedido) (Oferta, error) {
	if err := ctx.Err(); err != nil {
		return Oferta{}, fmt.Errorf("%s: %w", a.Nome(), err)
	}
	preco, prazo, err := a.sdk.CalcPrecoPrazo(p.CEPOrigem, p.CEPDestino, p.PesoGramas)
	if err != nil {
		return Oferta{}, fmt.Errorf("%s: %w", a.Nome(), err)
	}
	reais, centavos, ok := strings.Cut(preco, ",")
	if !ok {
		return Oferta{}, fmt.Errorf("%s: preco fora do formato: %q", a.Nome(), preco)
	}
	r, err1 := strconv.ParseInt(reais, 10, 64)
	c, err2 := strconv.ParseInt(centavos, 10, 64)
	d, err3 := strconv.Atoi(prazo)
	if err := errors.Join(err1, err2, err3); err != nil {
		return Oferta{}, fmt.Errorf("%s: resposta ilegivel: %w", a.Nome(), err)
	}
	return Oferta{a.Nome(), r*100 + c, time.Duration(d) * 24 * time.Hour}, nil
}

// ---------- Adaptador B: struct com reais (float) e horas ----------

type AdaptadorRapidoLog struct{ sdk legado.RapidoLog }

func NovoRapidoLog(atraso time.Duration) AdaptadorRapidoLog {
	return AdaptadorRapidoLog{legado.RapidoLog{Atraso: atraso}}
}

func (AdaptadorRapidoLog) Nome() string { return "rapido-log" }

func (a AdaptadorRapidoLog) Cotar(ctx context.Context, p Pedido) (Oferta, error) {
	if err := ctx.Err(); err != nil {
		return Oferta{}, fmt.Errorf("%s: %w", a.Nome(), err)
	}
	type res struct {
		r   legado.RespCotacao
		err error
	}
	ch := make(chan res, 1)
	go func() {
		r, err := a.sdk.Cotacao(legado.ReqCotacao{Origem: p.CEPOrigem, Destino: p.CEPDestino, PesoKg: float64(p.PesoGramas) / 1000})
		ch <- res{r, err}
	}()
	select {
	case <-ctx.Done():
		return Oferta{}, fmt.Errorf("%s: %w", a.Nome(), ctx.Err())
	case x := <-ch:
		if x.err != nil {
			return Oferta{}, fmt.Errorf("%s: %w", a.Nome(), x.err)
		}
		return Oferta{a.Nome(), int64(x.r.ValorTotal*100 + 0.5), time.Duration(x.r.PrazoHoras) * time.Hour}, nil
	}
}

// ---------- Adaptador C: callback -> retorno sincrono ----------

type AdaptadorMoto struct{ sdk legado.MotoExpress }

func NovaMoto(foraDoAr bool) AdaptadorMoto {
	return AdaptadorMoto{legado.MotoExpress{Fora: foraDoAr}}
}

func (AdaptadorMoto) Nome() string { return "moto-express" }

func (a AdaptadorMoto) Cotar(ctx context.Context, p Pedido) (Oferta, error) {
	if err := ctx.Err(); err != nil {
		return Oferta{}, fmt.Errorf("%s: %w", a.Nome(), err)
	}
	type res struct {
		centavos int64
		dias     int
		erro     string
	}
	ch := make(chan res, 1)
	a.sdk.Consultar(p.CEPOrigem, p.CEPDestino, float64(p.PesoGramas)/1000, func(c int64, d int, e string) {
		ch <- res{c, d, e}
	})
	select {
	case <-ctx.Done():
		return Oferta{}, fmt.Errorf("%s: %w", a.Nome(), ctx.Err())
	case x := <-ch:
		if x.erro != "" {
			return Oferta{}, fmt.Errorf("%s: %s", a.Nome(), x.erro)
		}
		return Oferta{a.Nome(), x.centavos, time.Duration(x.dias) * 24 * time.Hour}, nil
	}
}

// ---------- Cliente: usa so a interface ----------

// Cotar consulta todas em paralelo e devolve as ofertas ordenadas por preco,
// junto com os erros das que falharam (uma falha nao derruba as outras).
// Cada goroutine escreve na SUA posicao, entao nao precisa de trava e a
// ordem dos erros e a ordem da lista de transportadoras.
func Cotar(ctx context.Context, ts []Transportadora, p Pedido) ([]Oferta, error) {
	ofertas := make([]Oferta, len(ts))
	erros := make([]error, len(ts))
	var wg sync.WaitGroup
	for i, t := range ts {
		wg.Go(func() {
			ofertas[i], erros[i] = t.Cotar(ctx, p)
		})
	}
	wg.Wait()

	var ok []Oferta
	for i, o := range ofertas {
		if erros[i] == nil {
			ok = append(ok, o)
		}
	}
	slices.SortFunc(ok, func(a, b Oferta) int { return cmp.Compare(a.Centavos, b.Centavos) })
	return ok, errors.Join(erros...)
}
