// Package legado simula tres SDKs de transportadoras, cada um com seu jeito.
package legado

import (
	"errors"
	"fmt"
	"time"
)

// --- Transportadora A: tudo e texto, virgula decimal ---

type CorreiosLike struct{}

func (CorreiosLike) CalcPrecoPrazo(cepOrigem, cepDestino string, pesoGramas int) (preco, prazoDias string, err error) {
	if len(cepDestino) != 8 {
		return "", "", errors.New("CEP invalido")
	}
	centavos := 1500 + pesoGramas/10
	return fmt.Sprintf("%d,%02d", centavos/100, centavos%100), "3", nil
}

// --- Transportadora B: struct de requisicao e resposta, valor em reais (float), prazo em horas ---

type ReqCotacao struct {
	Origem, Destino string
	PesoKg          float64
}

type RespCotacao struct {
	ValorTotal float64
	PrazoHoras int
}

type RapidoLog struct{ Atraso time.Duration }

func (r RapidoLog) Cotacao(req ReqCotacao) (RespCotacao, error) {
	time.Sleep(r.Atraso)
	return RespCotacao{ValorTotal: 12.90 + req.PesoKg*3.10, PrazoHoras: 48}, nil
}

// --- Transportadora C: assincrona, devolve por callback, valor em centavos, erro em texto ---

type MotoExpress struct{ Fora bool }

func (m MotoExpress) Consultar(origem, destino string, kg float64, cb func(centavos int64, dias int, erro string)) {
	go func() {
		if m.Fora {
			cb(0, 0, "servico fora do ar")
			return
		}
		cb(int64(2000+kg*150), 1, "")
	}()
}
