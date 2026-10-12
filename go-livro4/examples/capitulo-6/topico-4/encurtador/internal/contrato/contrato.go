// Package contrato define o formato JSON da API de links curtos.
package contrato

import "time"

// PedidoCriar é o corpo de POST /links.
type PedidoCriar struct {
	URL string `json:"url"`
}

// RespostaCriar é a resposta de POST /links quando o link é criado.
type RespostaCriar struct {
	Codigo string `json:"codigo"`
	Curto  string `json:"curto"`
}

// Link é a resposta de GET /links/{codigo}.
type Link struct {
	Codigo  string     `json:"codigo"`
	Destino string     `json:"destino"`
	Criado  time.Time  `json:"criado"`
	Expira  *time.Time `json:"expira,omitempty"`
	Acessos int        `json:"acessos"`
}

// Erro é o corpo de qualquer resposta de erro.
type Erro struct {
	Mensagem string `json:"erro"`
}
