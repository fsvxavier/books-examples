package main

import "testing"

func TestMarcarRevelaOCampo(t *testing.T) {
	d := NovoDialogoPerfil()
	if d.NomeCao.Visivel {
		t.Fatal("o campo deveria comecar oculto")
	}
	d.TemCachorro.Alternar()
	if !d.NomeCao.Visivel {
		t.Fatal("marcar deveria revelar o campo")
	}
	d.TemCachorro.Alternar()
	if d.NomeCao.Visivel {
		t.Fatal("desmarcar deveria ocultar o campo")
	}
}

func TestSalvarSoComTudoPreenchido(t *testing.T) {
	d := NovoDialogoPerfil()
	d.Salvar.Clicar() // desabilitado: nao faz nada
	if len(d.Saida) != 0 {
		t.Fatal("salvou com o botao desabilitado")
	}
	d.NomeCliente.Digitar("Ana")
	if !d.Salvar.Habilitado {
		t.Fatal("sem cachorro, so o nome basta")
	}
	d.TemCachorro.Alternar()
	if d.Salvar.Habilitado {
		t.Fatal("com cachorro marcado e sem nome do cachorro, nao pode salvar")
	}
	d.NomeCao.Digitar("Rex")
	d.Salvar.Clicar()
	if len(d.Saida) != 1 {
		t.Fatalf("saida = %v", d.Saida)
	}
}

func TestDesmarcarLimpaOTextoDoCao(t *testing.T) {
	d := NovoDialogoPerfil()
	d.TemCachorro.Alternar()
	d.NomeCao.Digitar("Rex")
	d.TemCachorro.Alternar()
	if d.NomeCao.Texto != "" {
		t.Fatalf("texto = %q", d.NomeCao.Texto)
	}
}

// Os componentes funcionam SEM mediador: so nao fazem nada alem do proprio estado.
func TestComponenteSemMediador(t *testing.T) {
	c := NovaCaixa("x")
	c.Alternar()
	if !c.Marcada {
		t.Fatal("a caixa deveria mudar o proprio estado")
	}
}

// Um mediador falso prova que o componente notifica, e so isso.
type gravador struct{ eventos []string }

func (g *gravador) Notificar(o Componente, e string) { g.eventos = append(g.eventos, o.Nome()+":"+e) }

func TestComponenteNotificaOMediador(t *testing.T) {
	g := &gravador{}
	c := NovoCampo("nome")
	c.DefinirMediador(g)
	c.Digitar("a")
	b := NovoBotao("ok")
	b.DefinirMediador(g)
	b.Clicar()
	if len(g.eventos) != 2 || g.eventos[0] != "nome:digitou" || g.eventos[1] != "ok:clicou" {
		t.Fatalf("eventos = %v", g.eventos)
	}
}
