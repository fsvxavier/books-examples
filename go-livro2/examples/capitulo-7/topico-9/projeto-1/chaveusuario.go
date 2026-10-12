package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ---------- Identidade no contexto (chave de tipo proprio, Capitulo 12) ----------

type chaveUsuario struct{}

type Usuario struct {
	ID    int
	Admin bool
}

func ComUsuario(ctx context.Context, u Usuario) context.Context {
	return context.WithValue(ctx, chaveUsuario{}, u)
}

func UsuarioDe(ctx context.Context) (Usuario, bool) {
	u, ok := ctx.Value(chaveUsuario{}).(Usuario)
	return u, ok
}

// ---------- Proxy de PROTECAO: so o dono do perfil ou um admin ----------

type ProxyProtecao struct{ prox Perfis }

func Proteger(p Perfis) *ProxyProtecao { return &ProxyProtecao{p} }

func (p *ProxyProtecao) autorizar(ctx context.Context, id int) error {
	u, ok := UsuarioDe(ctx)
	if !ok {
		return fmt.Errorf("%w: sem identidade", ErrProibido)
	}
	if u.Admin || u.ID == id {
		return nil
	}
	return fmt.Errorf("%w: usuario %d nao pode acessar o perfil %d", ErrProibido, u.ID, id)
}

func (p *ProxyProtecao) Obter(ctx context.Context, id int) (Perfil, error) {
	if err := p.autorizar(ctx, id); err != nil {
		return Perfil{}, err
	}
	return p.prox.Obter(ctx, id)
}

func (p *ProxyProtecao) Atualizar(ctx context.Context, pf Perfil) error {
	if err := p.autorizar(ctx, pf.ID); err != nil {
		return err
	}
	return p.prox.Atualizar(ctx, pf)
}

// ---------- Proxy de LOG: registra cada chamada ----------

type ProxyLog struct {
	prox   Perfis
	nome   string
	mu     sync.Mutex
	Linhas []string
}

func ComLog(nome string, p Perfis) *ProxyLog { return &ProxyLog{prox: p, nome: nome} }

func (l *ProxyLog) registrar(f string, a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Linhas = append(l.Linhas, l.nome+": "+fmt.Sprintf(f, a...))
}

func (l *ProxyLog) Obter(ctx context.Context, id int) (Perfil, error) {
	p, err := l.prox.Obter(ctx, id)
	l.registrar("Obter(%d) -> %s", id, resumo(err))
	return p, err
}

func (l *ProxyLog) Atualizar(ctx context.Context, pf Perfil) error {
	err := l.prox.Atualizar(ctx, pf)
	l.registrar("Atualizar(%d) -> %s", pf.ID, resumo(err))
	return err
}

func resumo(err error) string {
	if err == nil {
		return "ok"
	}
	return strings.SplitN(err.Error(), ":", 2)[0]
}
