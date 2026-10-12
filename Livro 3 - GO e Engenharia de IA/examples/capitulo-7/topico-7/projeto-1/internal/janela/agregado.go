// Package janela implementa agregados em janela deslizante por entidade,
// em três versões com custos diferentes.
package janela

import "time"

// Agregado é o resultado de uma janela: quantos eventos e quanto somaram.
type Agregado struct {
	Contagem int64
	Centav   int64
}

// Janela guarda eventos por entidade e responde "quanto houve nos últimos
// D?". O tempo é sempre passado por parâmetro: assim o mesmo código serve
// para reprocessar o passado (treino) e para o presente (serviço), e os
// testes não dependem do relógio.
type Janela interface {
	Adicionar(entidade string, quando time.Time, centav int64)
	Consultar(entidade string, agora time.Time) Agregado
}
