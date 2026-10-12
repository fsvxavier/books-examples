package relatorio

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
)

type Linha struct {
	Produto string
	Total   int64
}

// Exportador escreve linhas em um destino. Fechar finaliza o formato (rodape, buffer).
type Exportador interface {
	Escrever(l Linha) error
	Fechar() error
}

// Fabrica recebe o destino: o servico nao sabe se e arquivo, rede ou buffer de teste.
type Fabrica func(w io.Writer) Exportador

type csvExp struct{ w *csv.Writer }

func NovoCSV(w io.Writer) Exportador {
	c := csv.NewWriter(w)
	c.Write([]string{"produto", "total"})
	return &csvExp{c}
}

func (e *csvExp) Escrever(l Linha) error {
	return e.w.Write([]string{l.Produto, fmt.Sprint(l.Total)})
}

func (e *csvExp) Fechar() error { e.w.Flush(); return e.w.Error() }

type jsonExp struct {
	enc *json.Encoder
	w   io.Writer
	n   int
}

func NovoJSONL(w io.Writer) Exportador { return &jsonExp{enc: json.NewEncoder(w), w: w} }

func (e *jsonExp) Escrever(l Linha) error { e.n++; return e.enc.Encode(l) }
func (e *jsonExp) Fechar() error          { return nil }

// Servico depende da FABRICA, nao de um formato.
type Servico struct{ Nova Fabrica }

func (s Servico) Exportar(w io.Writer, linhas []Linha) (err error) {
	exp := s.Nova(w)
	defer func() {
		if e := exp.Fechar(); err == nil { // Fechar sempre, mesmo em erro
			err = e
		}
	}()
	for _, l := range linhas {
		if err := exp.Escrever(l); err != nil {
			return fmt.Errorf("escrever %q: %w", l.Produto, err)
		}
	}
	return nil
}
