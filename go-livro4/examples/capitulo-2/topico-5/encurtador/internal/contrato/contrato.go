// Package contrato define o formato JSON da API de links curtos.
// Servidor e clientes importam este pacote, e só ele: ninguém precisa
// conhecer a implementação do outro lado.
package contrato

// PedidoCriar é o corpo de POST /links.
type PedidoCriar struct {
	URL string `json:"url"`
}

// RespostaCriar é a resposta de POST /links quando o link é criado.
type RespostaCriar struct {
	Codigo string `json:"codigo"`
	Curto  string `json:"curto"`
}

// Erro é o corpo de qualquer resposta de erro.
type Erro struct {
	Mensagem string `json:"erro"`
}
