package main

import (
	"context"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// ---------- Proxy de CACHE: TTL, relogio injetado, invalidacao na escrita, sem estampida ----------

type entrada struct {
	perfil Perfil
	expira time.Time
}

type ProxyCache struct {
	prox  Perfis
	ttl   time.Duration
	agora func() time.Time

	mu    sync.Mutex
	itens map[int]entrada
	voo   singleflight.Group

	Acertos, Falhas int // so para o experimento (protegidos por mu)
}

func ComCache(p Perfis, ttl time.Duration, agora func() time.Time) *ProxyCache {
	return &ProxyCache{prox: p, ttl: ttl, agora: agora, itens: map[int]entrada{}}
}

func (c *ProxyCache) Obter(ctx context.Context, id int) (Perfil, error) {
	c.mu.Lock()
	if e, ok := c.itens[id]; ok && c.agora().Before(e.expira) {
		c.Acertos++
		c.mu.Unlock()
		return e.perfil, nil
	}
	c.Falhas++
	c.mu.Unlock()

	// Varios pedidos simultaneos para o MESMO id viram UMA consulta ao servico real.
	v, err, _ := c.voo.Do(strconv.Itoa(id), func() (any, error) {
		p, err := c.prox.Obter(ctx, id)
		if err != nil {
			return Perfil{}, err // erros nao entram no cache
		}
		c.mu.Lock()
		c.itens[id] = entrada{p, c.agora().Add(c.ttl)}
		c.mu.Unlock()
		return p, nil
	})
	if err != nil {
		return Perfil{}, err
	}
	return v.(Perfil), nil
}

// Atualizar escreve no servico real e INVALIDA a entrada: sem isso, o cache serviria dado velho.
func (c *ProxyCache) Atualizar(ctx context.Context, p Perfil) error {
	if err := c.prox.Atualizar(ctx, p); err != nil {
		return err
	}
	c.mu.Lock()
	delete(c.itens, p.ID)
	c.mu.Unlock()
	return nil
}
