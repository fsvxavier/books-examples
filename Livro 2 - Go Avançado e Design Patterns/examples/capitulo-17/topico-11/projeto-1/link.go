package e09

import (
	"context"
	"errors"
)

// Link e Mensagem são duas entidades do encurtador (a caixa de saída do Capítulo 45).
type Link struct{ Codigo, Destino string }

// Mensagem é o que a caixa de saída guarda.
type Mensagem struct{ ID, Topico string }

// CriarComRepositoriosGenericos faz o que o Capítulo 45 exige: gravar o link E a mensagem.
// Com dois repositórios genéricos independentes, não há como fazer os dois valerem juntos.
func CriarComRepositoriosGenericos(ctx context.Context, links Repositorio[Link, string], caixa Repositorio[Mensagem, string], l Link) error {
	if err := links.Salvar(ctx, l.Codigo, l); err != nil {
		return err
	}
	// Se a caixa falhar aqui, o link já está gravado e a mensagem nunca sairá.
	return caixa.Salvar(ctx, "msg-"+l.Codigo, Mensagem{ID: "msg-" + l.Codigo, Topico: "link_criado"})
}

// Gravador é uma porta ESPECÍFICA do caso de uso: sabe que link e mensagem andam juntos.
type Gravador interface {
	GravarLinkEMensagem(ctx context.Context, l Link, m Mensagem) error
}

// GravadorMemoria cumpre a porta específica com uma "transação" simples:
// valida tudo antes de gravar, e só então grava os dois (ou nenhum).
type GravadorMemoria struct {
	Links  *Memoria[Link, string]
	Caixa  *Memoria[Mensagem, string]
	Falhar func() error
}

// GravarLinkEMensagem grava os dois ou nenhum.
func (g *GravadorMemoria) GravarLinkEMensagem(ctx context.Context, l Link, m Mensagem) error {
	if g.Falhar != nil {
		if err := g.Falhar(); err != nil {
			return err // nada foi gravado
		}
	}
	return errors.Join(g.Links.Salvar(ctx, l.Codigo, l), g.Caixa.Salvar(ctx, m.ID, m))
}
