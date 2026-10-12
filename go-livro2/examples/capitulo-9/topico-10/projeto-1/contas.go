package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

// ---------- Receptor: contas (so dados e regras; nao sabe de comandos) ----------

type Contas struct {
	mu     sync.Mutex
	saldos map[string]int64
}

func NovasContas() *Contas { return &Contas{saldos: map[string]int64{}} }

func (c *Contas) Saldo(conta string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.saldos[conta]
}

func (c *Contas) Snapshot() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make(map[string]int64, len(c.saldos))
	for k, v := range c.saldos {
		cp[k] = v
	}
	return cp
}

var ErrSaldoInsuficiente = errors.New("saldo insuficiente")

// ---------- Comando serializavel: tudo que e preciso esta em DADOS (nome + argumentos) ----------

type Comando interface {
	Tipo() string
	Executar(*Contas) error
}

type Depositar struct {
	Conta string `json:"conta"`
	Valor int64  `json:"valor"`
}

func (Depositar) Tipo() string { return "depositar" }
func (d Depositar) Executar(c *Contas) error {
	if d.Valor <= 0 {
		return fmt.Errorf("valor invalido: %d", d.Valor)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.saldos[d.Conta] += d.Valor
	return nil
}

type Sacar struct {
	Conta string `json:"conta"`
	Valor int64  `json:"valor"`
}

func (Sacar) Tipo() string { return "sacar" }
func (s Sacar) Executar(c *Contas) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.saldos[s.Conta] < s.Valor {
		return fmt.Errorf("%w: %s tem %d, pediu %d", ErrSaldoInsuficiente, s.Conta, c.saldos[s.Conta], s.Valor)
	}
	c.saldos[s.Conta] -= s.Valor
	return nil
}

// ---------- Registro de decodificadores (o Registry do Capitulo 27) ----------

type decodificador func(json.RawMessage) (Comando, error)

var registro = map[string]decodificador{
	"depositar": decodificar[Depositar],
	"sacar":     decodificar[Sacar],
}

func decodificar[T Comando](dados json.RawMessage) (Comando, error) {
	var c T
	if err := json.Unmarshal(dados, &c); err != nil {
		return nil, err
	}
	return c, nil
}

// Envelope e o formato gravado: o ID permite ignorar repeticoes (idempotencia).
type Envelope struct {
	ID    string          `json:"id"`
	Tipo  string          `json:"tipo"`
	Dados json.RawMessage `json:"dados"`
}

func Empacotar(id string, c Comando) (Envelope, error) {
	dados, err := json.Marshal(c)
	return Envelope{ID: id, Tipo: c.Tipo(), Dados: dados}, err
}

func (e Envelope) Comando() (Comando, error) {
	dec, ok := registro[e.Tipo]
	if !ok {
		return nil, fmt.Errorf("tipo de comando desconhecido: %q", e.Tipo)
	}
	return dec(e.Dados)
}

// ---------- Diario (journal): grava cada comando ANTES de executar ----------

type Diario struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (d *Diario) Gravar(e Envelope) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return json.NewEncoder(&d.buf).Encode(e)
}

func (d *Diario) Bytes() []byte {
	d.mu.Lock()
	defer d.mu.Unlock()
	return bytes.Clone(d.buf.Bytes())
}

// Processador executa envelopes UMA vez por ID.
type Processador struct {
	Contas *Contas
	Diario *Diario
	mu     sync.Mutex
	vistos map[string]bool
	Falhas []string
}

func NovoProcessador(c *Contas, d *Diario) *Processador {
	return &Processador{Contas: c, Diario: d, vistos: map[string]bool{}}
}

func (p *Processador) Processar(e Envelope) error {
	p.mu.Lock()
	if p.vistos[e.ID] {
		p.mu.Unlock()
		return nil // repeticao: ja executado, ignora (idempotente)
	}
	p.vistos[e.ID] = true
	p.mu.Unlock()

	cmd, err := e.Comando()
	if err != nil {
		return fmt.Errorf("comando %s: %w", e.ID, err)
	}
	if p.Diario != nil {
		if err := p.Diario.Gravar(e); err != nil {
			return err
		}
	}
	if err := cmd.Executar(p.Contas); err != nil {
		p.mu.Lock()
		p.Falhas = append(p.Falhas, e.ID)
		p.mu.Unlock()
		return fmt.Errorf("comando %s (%s): %w", e.ID, e.Tipo, err)
	}
	return nil
}

// Reproduzir le o diario e executa tudo de novo em contas NOVAS: reconstroi o estado.
func Reproduzir(r io.Reader) (*Contas, error) {
	contas := NovasContas()
	p := NovoProcessador(contas, nil)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		var e Envelope
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			return nil, err
		}
		if err := p.Processar(e); err != nil && !errors.Is(err, ErrSaldoInsuficiente) {
			return nil, err
		}
	}
	return contas, sc.Err()
}

// ---------- Fila: producao e execucao desacopladas (execucao adiada) ----------

type Fila struct {
	ch   chan Envelope
	proc *Processador
	wg   sync.WaitGroup
}

func NovaFila(p *Processador, buffer int) *Fila {
	f := &Fila{ch: make(chan Envelope, buffer), proc: p}
	f.wg.Go(func() {
		for e := range f.ch {
			f.proc.Processar(e) // os erros ficam registrados em Falhas
		}
	})
	return f
}

func (f *Fila) Enviar(e Envelope) { f.ch <- e }
func (f *Fila) Encerrar()         { close(f.ch); f.wg.Wait() }
