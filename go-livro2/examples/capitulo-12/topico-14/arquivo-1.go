package main

import "fmt"

// ---- Strategy: COMO se desbloqueia. Escolhida pelo usuário, nunca muda sozinha.

type Desbloqueio interface {
	Nome() string
	Verificar(credencial string) bool
}

type PIN struct{ Correto string }

func (PIN) Nome() string              { return "PIN" }
func (p PIN) Verificar(c string) bool { return c == p.Correto }

type Digital struct{ Cadastradas map[string]bool }

func (Digital) Nome() string              { return "digital" }
func (d Digital) Verificar(c string) bool { return d.Cadastradas[c] }

// ---- State: O QUE o botão faz AGORA. O próprio estado decide o próximo.

type Estado interface {
	Nome() string
	Energia(t *Telefone) string
	Digitar(t *Telefone, credencial string) string
}

type Telefone struct {
	estado      Estado
	Desbloqueio Desbloqueio // a estratégia: o telefone só a usa, não a troca
	bateria     int
	Tela        []string
}

func NovoTelefone(d Desbloqueio) *Telefone {
	return &Telefone{estado: &Bloqueado{}, Desbloqueio: d, bateria: 100}
}

func (t *Telefone) Energia() string                  { return t.estado.Energia(t) }
func (t *Telefone) Digitar(credencial string) string { return t.estado.Digitar(t, credencial) }
func (t *Telefone) Estado() string                   { return t.estado.Nome() }

func (t *Telefone) Bateria(n int) {
	t.bateria = n
	if n <= 5 {
		t.estado = BateriaFraca{}
	}
}

// Bloqueado tem DADOS: conta as tentativas erradas. Por isso é um ponteiro,
// e uma transição para ele cria um valor novo.
type Bloqueado struct{ Erros int }

func (*Bloqueado) Nome() string { return "bloqueado" }
func (b *Bloqueado) Energia(t *Telefone) string {
	return "tela acende, pede " + t.Desbloqueio.Nome()
}
func (b *Bloqueado) Digitar(t *Telefone, c string) string {
	if t.Desbloqueio.Verificar(c) {
		t.estado = Desbloqueado{}
		return "desbloqueado"
	}
	b.Erros++
	if b.Erros >= 3 {
		t.estado = Travado{}
		return "travado após 3 erros"
	}
	return fmt.Sprintf("errado (%d de 3)", b.Erros)
}

type Desbloqueado struct{}

func (Desbloqueado) Nome() string { return "desbloqueado" }
func (Desbloqueado) Energia(t *Telefone) string {
	t.estado = &Bloqueado{}
	return "tela apaga e o telefone bloqueia"
}
func (Desbloqueado) Digitar(*Telefone, string) string { return "digitou no teclado normal" }

type Travado struct{}

func (Travado) Nome() string                     { return "travado" }
func (Travado) Energia(*Telefone) string         { return "nada acontece" }
func (Travado) Digitar(*Telefone, string) string { return "recusado: telefone travado" }

type BateriaFraca struct{}

func (BateriaFraca) Nome() string { return "bateria fraca" }
func (BateriaFraca) Energia(t *Telefone) string {
	return "aviso de bateria, tela em modo econômico"
}
func (BateriaFraca) Digitar(t *Telefone, c string) string {
	if t.Desbloqueio.Verificar(c) {
		t.estado = Desbloqueado{}
		return "desbloqueado (modo econômico mantido)"
	}
	return "errado"
}

func passo(t *Telefone, acao string, f func() string) {
	r := f()
	fmt.Printf("  %-22s -> %-34s [%s]\n", acao, r, t.Estado())
}

func main() {
	fmt.Println("usuário escolheu PIN (a estratégia):")
	t := NovoTelefone(PIN{"1234"})
	passo(t, "botão de energia", t.Energia)
	passo(t, "digita 0000", func() string { return t.Digitar("0000") })
	passo(t, "digita 1234", func() string { return t.Digitar("1234") })
	passo(t, "botão de energia", t.Energia)

	fmt.Println("mesmo telefone, usuário troca a estratégia para digital:")
	t.Desbloqueio = Digital{map[string]bool{"dedo-A": true}}
	passo(t, "digita dedo-A", func() string { return t.Digitar("dedo-A") })

	fmt.Println("três erros travam (o ESTADO decide a transição):")
	u := NovoTelefone(PIN{"1234"})
	for range 3 {
		u.Digitar("0")
	}
	passo(u, "botão de energia", u.Energia)
	passo(u, "digita 1234", func() string { return u.Digitar("1234") })

	fmt.Println("bateria fraca muda o comportamento de TODOS os botões:")
	v := NovoTelefone(PIN{"1234"})
	v.Bateria(4)
	passo(v, "botão de energia", v.Energia)
}
