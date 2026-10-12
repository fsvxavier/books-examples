// Package banco expoe uma unica conexao simulada.
package banco

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Conexao e o que o resto do programa enxerga: so uma interface.
type Conexao interface {
	Consultar(sql string) string
	ID() int64
}

// pool NAO e exportado: fora deste pacote ninguem escreve banco.pool{}.
type pool struct{ id int64 }

var contador atomic.Int64

func (p *pool) Consultar(sql string) string { return fmt.Sprintf("pool %d executou %q", p.id, sql) }
func (p *pool) ID() int64                   { return p.id }

var instancia = sync.OnceValue(func() Conexao {
	return &pool{id: contador.Add(1)}
})

// Obter e o unico ponto de acesso: devolve sempre a mesma instancia.
func Obter() Conexao { return instancia() }

// PoolAberto e o contraexemplo: tipo exportado, qualquer um cria outro.
type PoolAberto struct{ ID int64 }
