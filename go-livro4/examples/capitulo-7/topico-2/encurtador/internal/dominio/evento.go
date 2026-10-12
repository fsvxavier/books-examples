package dominio

import "time"

// Evento é algo que aconteceu no domínio e que interessa a outras partes do
// sistema. O nome está no passado, porque um evento é um fato consumado.
type Evento interface {
	// Nome identifica o tipo do evento, por exemplo "link_criado".
	Nome() string
	// Quando devolve o instante em que o fato ocorreu.
	Quando() time.Time
}

// LinkCriado registra que um link foi criado.
type LinkCriado struct {
	Codigo  Codigo
	Destino Destino
	Expira  time.Time // zero quando o link não tem prazo
	Em      time.Time
}

// Nome implementa Evento.
func (LinkCriado) Nome() string { return "link_criado" }

// Quando implementa Evento.
func (e LinkCriado) Quando() time.Time { return e.Em }

// LinkAcessado registra que um link foi seguido.
type LinkAcessado struct {
	Codigo Codigo
	Em     time.Time
}

// Nome implementa Evento.
func (LinkAcessado) Nome() string { return "link_acessado" }

// Quando implementa Evento.
func (e LinkAcessado) Quando() time.Time { return e.Em }
