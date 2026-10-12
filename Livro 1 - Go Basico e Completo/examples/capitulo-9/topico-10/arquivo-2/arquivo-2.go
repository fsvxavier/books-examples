package main

import (
	"fmt"
	"sort"
	"sync"
)

// Hub: UMA goroutine é dona do mapa de clientes. Todo o resto conversa com ela por canais.
// Sem mutex, sem corrida: o estado nunca é compartilhado, só confinado.
type Hub struct {
	entrar  chan *Cliente
	sair    chan *Cliente
	enviar  chan string
	parar   chan struct{}
	pronto  sync.WaitGroup
	clients map[*Cliente]struct{}
}

type Cliente struct {
	Nome  string
	Caixa chan string
}

func NovoHub() *Hub {
	h := &Hub{
		entrar:  make(chan *Cliente),
		sair:    make(chan *Cliente),
		enviar:  make(chan string),
		parar:   make(chan struct{}),
		clients: make(map[*Cliente]struct{}),
	}
	h.pronto.Add(1)
	go h.laco()
	return h
}

func (h *Hub) laco() {
	defer h.pronto.Done()
	for {
		select {
		case c := <-h.entrar:
			h.clients[c] = struct{}{}
		case c := <-h.sair:
			delete(h.clients, c)
			close(c.Caixa)
		case msg := <-h.enviar:
			for c := range h.clients {
				select {
				case c.Caixa <- msg:
				default: // caixa cheia: descarta para não travar o hub inteiro
				}
			}
		case <-h.parar:
			for c := range h.clients {
				close(c.Caixa)
			}
			return
		}
	}
}

func (h *Hub) Parar() {
	close(h.parar)
	h.pronto.Wait()
}

func novoCliente(nome string) *Cliente { return &Cliente{nome, make(chan string, 10)} }

func main() {
	hub := NovoHub()
	ana, bia, caio := novoCliente("ana"), novoCliente("bia"), novoCliente("caio")
	hub.entrar <- ana
	hub.entrar <- bia
	hub.entrar <- caio

	hub.enviar <- "olá a todos"
	hub.sair <- bia
	hub.enviar <- "bia saiu"
	hub.Parar()

	for _, c := range []*Cliente{ana, bia, caio} {
		var msgs []string
		for m := range c.Caixa {
			msgs = append(msgs, m)
		}
		sort.Strings(msgs)
		fmt.Printf("%-5s recebeu %q\n", c.Nome, msgs)
	}
}
