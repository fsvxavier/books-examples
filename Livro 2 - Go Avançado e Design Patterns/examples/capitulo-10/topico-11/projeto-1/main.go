package main

import "fmt"

func estado(d *DialogoPerfil) string {
	return fmt.Sprintf("campo-cao visivel=%-5v salvar habilitado=%v", d.NomeCao.Visivel, d.Salvar.Habilitado)
}

// Reuso: a MESMA classe Caixa, em outro dialogo, com outro mediador e outra regra.
type DialogoTermos struct {
	Aceito *Caixa
	Enviar *Botao
}

func (d *DialogoTermos) Notificar(origem Componente, evento string) {
	if origem == Componente(d.Aceito) {
		d.Enviar.Habilitado = d.Aceito.Marcada
	}
}

func main() {
	d := NovoDialogoPerfil()
	fmt.Println("inicio:                        ", estado(d))

	d.TemCachorro.Alternar()
	fmt.Println("marcou 'tem cachorro':         ", estado(d))

	d.NomeCliente.Digitar("Ana")
	fmt.Println("digitou o nome do cliente:     ", estado(d))

	d.NomeCao.Digitar("Rex")
	fmt.Println("digitou o nome do cachorro:    ", estado(d))

	d.Salvar.Clicar()
	fmt.Println("clicou em salvar:              ", d.Saida)

	d.TemCachorro.Alternar()
	fmt.Println("desmarcou 'tem cachorro':      ", estado(d), "| texto do cao:", fmt.Sprintf("%q", d.NomeCao.Texto))

	d.Salvar.Clicar()
	fmt.Println("clicou em salvar de novo:      ", len(d.Saida), "registro(s)")

	// A mesma Caixa, reaproveitada com outro mediador.
	t := &DialogoTermos{Aceito: NovaCaixa("aceito os termos"), Enviar: NovoBotao("enviar")}
	t.Enviar.Habilitado = false
	t.Aceito.DefinirMediador(t)
	t.Aceito.Alternar()
	fmt.Println("dialogo de termos, marcou:     ", "enviar habilitado =", t.Enviar.Habilitado)
}
