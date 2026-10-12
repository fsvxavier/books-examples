package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrFormato    = errors.New("cabecalho de assinatura malformado")
	ErrAssinatura = errors.New("assinatura invalida")
	ErrAntigo     = errors.New("timestamp fora da tolerancia")
	ErrRepetido   = errors.New("evento ja processado")
)

// Assinar monta o cabecalho "t=<unix>,v1=<hmac>". O HMAC cobre o timestamp E o corpo,
// para que ninguem reaproveite uma assinatura antiga com outro horario.
func Assinar(segredo []byte, corpo []byte, agora time.Time) string {
	ts := strconv.FormatInt(agora.Unix(), 10)
	return fmt.Sprintf("t=%s,v1=%s", ts, calcular(segredo, ts, corpo))
}

func calcular(segredo []byte, ts string, corpo []byte) string {
	m := hmac.New(sha256.New, segredo)
	m.Write([]byte(ts))
	m.Write([]byte("."))
	m.Write(corpo)
	return hex.EncodeToString(m.Sum(nil))
}

type Verificador struct {
	Segredo    []byte
	Tolerancia time.Duration
	Agora      func() time.Time // injetavel: testes controlam o relogio

	mu     sync.Mutex
	vistos map[string]time.Time
}

func NovoVerificador(segredo []byte, tol time.Duration) *Verificador {
	return &Verificador{Segredo: segredo, Tolerancia: tol, Agora: time.Now, vistos: map[string]time.Time{}}
}

// Verificar confere formato, assinatura (em tempo constante), idade e repeticao.
func (v *Verificador) Verificar(cabecalho string, corpo []byte, idEvento string) error {
	var ts, assinatura string
	for _, p := range strings.Split(cabecalho, ",") {
		k, val, ok := strings.Cut(p, "=")
		if !ok {
			return ErrFormato
		}
		switch k {
		case "t":
			ts = val
		case "v1":
			assinatura = val
		}
	}
	unix, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || assinatura == "" {
		return ErrFormato
	}
	esperado := calcular(v.Segredo, ts, corpo)
	if !hmac.Equal([]byte(esperado), []byte(assinatura)) {
		return ErrAssinatura
	}
	agora := v.Agora()
	if d := agora.Sub(time.Unix(unix, 0)); d > v.Tolerancia || d < -v.Tolerancia {
		return ErrAntigo
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	for id, quando := range v.vistos { // limpa o que ja expirou
		if agora.Sub(quando) > v.Tolerancia {
			delete(v.vistos, id)
		}
	}
	if _, ja := v.vistos[idEvento]; ja {
		return ErrRepetido
	}
	v.vistos[idEvento] = agora
	return nil
}
