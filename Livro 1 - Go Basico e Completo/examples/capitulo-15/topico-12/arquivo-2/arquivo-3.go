package pedido

import "sync"

// RepoMemoria guarda os pedidos em memória. Seguro para uso concorrente.
type RepoMemoria struct {
	mu      sync.Mutex
	pedidos map[int]Pedido
	proximo int
}

func NovoRepoMemoria() *RepoMemoria {
	return &RepoMemoria{pedidos: make(map[int]Pedido), proximo: 1}
}

func (r *RepoMemoria) Salvar(p Pedido) (Pedido, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p.ID == 0 {
		p.ID = r.proximo
		r.proximo++
	}
	r.pedidos[p.ID] = p
	return p, nil
}

func (r *RepoMemoria) Buscar(id int) (Pedido, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.pedidos[id]
	if !ok {
		return Pedido{}, ErrNaoEncontrado
	}
	return p, nil
}
