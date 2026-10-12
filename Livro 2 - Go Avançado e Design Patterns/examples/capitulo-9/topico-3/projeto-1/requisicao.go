package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Requisicao e o OBJETO que viaja pela cadeia (a forma mais flexivel segundo o catalogo).
type Requisicao struct {
	Usuario string
	Senha   string
	IP      string
	Admin   bool
	Corpo   string

	Autenticado bool // preenchido pelos manipuladores
}

var (
	ErrNaoAutenticado = errors.New("nao autenticado")
	ErrBloqueado      = errors.New("muitas tentativas falhas deste IP")
	ErrInvalida       = errors.New("requisicao invalida")
)

// Manipulador e a interface comum: um metodo, e o elo para o proximo.
type Manipulador interface {
	Tratar(ctx context.Context, r *Requisicao) error
	DefinirProximo(Manipulador) Manipulador // devolve o proximo, para encadear em uma linha
}

// base elimina o codigo repetido: guarda o proximo e repassa por padrao.
type base struct{ prox Manipulador }

func (b *base) DefinirProximo(m Manipulador) Manipulador { b.prox = m; return m }

func (b *base) repassar(ctx context.Context, r *Requisicao) error {
	if b.prox == nil {
		return nil // fim da cadeia: a requisicao passou por todos
	}
	return b.prox.Tratar(ctx, r)
}

// ---------- Manipuladores concretos ----------

// Autenticacao: se as credenciais falham, NAO repassa (interrompe a cadeia).
type Autenticacao struct {
	base
	Usuarios map[string]string
}

func (a *Autenticacao) Tratar(ctx context.Context, r *Requisicao) error {
	if senha, ok := a.Usuarios[r.Usuario]; !ok || senha != r.Senha {
		return fmt.Errorf("%w: usuario %q", ErrNaoAutenticado, r.Usuario)
	}
	r.Autenticado = true
	return a.repassar(ctx, r)
}

// LimiteDeFalhas: bloqueia IPs com falhas demais. Registra as falhas que o proximo elo devolver.
type LimiteDeFalhas struct {
	base
	Maximo int
	mu     sync.Mutex
	falhas map[string]int
}

func NovoLimite(max int) *LimiteDeFalhas {
	return &LimiteDeFalhas{Maximo: max, falhas: map[string]int{}}
}

func (l *LimiteDeFalhas) Tratar(ctx context.Context, r *Requisicao) error {
	l.mu.Lock()
	n := l.falhas[r.IP]
	l.mu.Unlock()
	if n >= l.Maximo {
		return fmt.Errorf("%w: %s (%d falhas)", ErrBloqueado, r.IP, n)
	}
	err := l.repassar(ctx, r)
	if errors.Is(err, ErrNaoAutenticado) {
		l.mu.Lock()
		l.falhas[r.IP]++
		l.mu.Unlock()
	}
	return err
}

// Sanitizacao: normaliza e valida o corpo antes de entregar ao sistema.
type Sanitizacao struct{ base }

func (s *Sanitizacao) Tratar(ctx context.Context, r *Requisicao) error {
	r.Corpo = strings.TrimSpace(r.Corpo)
	if r.Corpo == "" || strings.ContainsAny(r.Corpo, "<>") {
		return fmt.Errorf("%w: corpo vazio ou com marcacao", ErrInvalida)
	}
	return s.repassar(ctx, r)
}

// Cache: se ja temos a resposta, ENCERRA a cadeia com sucesso (nao repassa).
type Cache struct {
	base
	mu        sync.Mutex
	respostas map[string]string
	Acertos   int
}

func NovoCache() *Cache { return &Cache{respostas: map[string]string{}} }

func (c *Cache) Tratar(ctx context.Context, r *Requisicao) error {
	c.mu.Lock()
	if _, ok := c.respostas[r.Corpo]; ok {
		c.Acertos++
		c.mu.Unlock()
		return nil // tratada aqui: ninguem mais precisa agir
	}
	c.mu.Unlock()
	if err := c.repassar(ctx, r); err != nil {
		return err
	}
	c.mu.Lock()
	c.respostas[r.Corpo] = "ok"
	c.mu.Unlock()
	return nil
}

// Pedidos e o "sistema de pedidos" no fim da cadeia.
type Pedidos struct {
	base
	mu      sync.Mutex
	Criados []string
}

func (p *Pedidos) Tratar(ctx context.Context, r *Requisicao) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Criados = append(p.Criados, r.Usuario+": "+r.Corpo)
	return nil
}
