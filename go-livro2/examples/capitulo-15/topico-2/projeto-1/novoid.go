package e01

import "sync"

// Versão com estado global: dependências e contadores escondidos em variáveis de pacote.
var (
	config    = map[string]string{"moeda": "BRL"}
	proximoID int
)

func NovoID() int {
	proximoID++
	return proximoID
}

func Moeda() string { return config["moeda"] }

// Versão com dependência explícita: o estado vive numa instância.
type Gerador struct {
	mu     sync.Mutex
	ultimo int
}

func (g *Gerador) NovoID() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ultimo++
	return g.ultimo
}

type Loja struct {
	Moeda string
	IDs   *Gerador
}

func NovaLoja(moeda string, ids *Gerador) *Loja {
	return &Loja{Moeda: moeda, IDs: ids}
}
