package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Config é IMUTÁVEL depois de publicada: quem quer mudar algo cria uma nova.
type Config struct {
	Versao         int
	LimiteTaxa     int
	ModoManutencao bool
}

// Servico lê a configuração em todo pedido, sem travar, e aceita recarga a qualquer momento.
type Servico struct {
	cfg atomic.Pointer[Config]
}

func NovoServico(inicial Config) *Servico {
	s := &Servico{}
	s.cfg.Store(&inicial)
	return s
}

func (s *Servico) Config() *Config { return s.cfg.Load() }

// Recarregar publica uma nova versão. Os leitores em andamento continuam com a versão antiga, inteira e consistente.
func (s *Servico) Recarregar(muda func(c *Config)) {
	for {
		antiga := s.cfg.Load()
		nova := *antiga // cópia
		muda(&nova)
		nova.Versao = antiga.Versao + 1
		if s.cfg.CompareAndSwap(antiga, &nova) {
			return
		}
		// outra recarga publicou antes: tenta de novo sobre a versão mais recente
	}
}

func (s *Servico) Atender() (versao, limite int) {
	c := s.Config() // UMA leitura: versão e limite vêm da mesma Config, sempre coerentes
	return c.Versao, c.LimiteTaxa
}

func main() {
	s := NovoServico(Config{Versao: 1, LimiteTaxa: 10})

	var incoerencias atomic.Int64
	var atendidos atomic.Int64
	parar := make(chan struct{})

	var leitores sync.WaitGroup
	for range 4 {
		leitores.Go(func() {
			for {
				select {
				case <-parar:
					return
				default:
				}
				v, l := s.Atender()
				atendidos.Add(1)
				if l != v*10 { // regra: limite == versão * 10; qualquer outra combinação seria uma leitura "rasgada"
					incoerencias.Add(1)
				}
			}
		})
	}

	var recarregadores sync.WaitGroup
	for range 3 {
		recarregadores.Go(func() {
			for range 50 {
				s.Recarregar(func(c *Config) { c.LimiteTaxa = (c.Versao + 1) * 10 })
				time.Sleep(100 * time.Microsecond)
			}
		})
	}
	recarregadores.Wait()
	close(parar)
	leitores.Wait()

	fmt.Println("versão final:", s.Config().Versao)
	fmt.Println("leituras incoerentes:", incoerencias.Load(), "| houve leituras:", atendidos.Load() > 0)
}
