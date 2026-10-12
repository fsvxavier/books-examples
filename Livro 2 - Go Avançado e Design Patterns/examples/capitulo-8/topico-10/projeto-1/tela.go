package main

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Tela e o destino do desenho (aqui, um contador para podermos testar).
type Tela interface {
	Pintar(nome, cor string, texturaBytes int, x, y float64)
}

// ---------- Flyweight: estado intrinseco, imutavel ----------

type TipoArvore struct {
	nome    string
	cor     string
	textura []byte
}

func (t *TipoArvore) Nome() string { return t.nome }

// Desenhar recebe o estado extrinseco (x, y) como parametro.
func (t *TipoArvore) Desenhar(tela Tela, x, y float64) {
	tela.Pintar(t.nome, t.cor, len(t.textura), x, y)
}

// ---------- Fabrica de flyweights ----------

type FabricaTipos struct {
	mu    sync.Mutex
	tipos map[chaveTipo]*TipoArvore
}

type chaveTipo struct{ nome, cor string }

func NovaFabricaTipos() *FabricaTipos { return &FabricaTipos{tipos: map[chaveTipo]*TipoArvore{}} }

func (f *FabricaTipos) Obter(nome, cor string, textura []byte) *TipoArvore {
	f.mu.Lock()
	defer f.mu.Unlock()
	k := chaveTipo{nome, cor}
	if t, ok := f.tipos[k]; ok {
		return t
	}
	t := &TipoArvore{nome: nome, cor: cor, textura: append([]byte(nil), textura...)}
	f.tipos[k] = t
	return t
}

func (f *FabricaTipos) Quantidade() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.tipos)
}

// ---------- Contexto: estado extrinseco + referencia ao flyweight ----------

type Arvore struct {
	X, Y float64
	Tipo *TipoArvore
}

type Floresta struct {
	fabrica *FabricaTipos
	arvores []Arvore
}

func NovaFloresta(f *FabricaTipos) *Floresta { return &Floresta{fabrica: f} }

func (fl *Floresta) Plantar(x, y float64, nome, cor string, textura []byte) {
	fl.arvores = append(fl.arvores, Arvore{x, y, fl.fabrica.Obter(nome, cor, textura)})
}

func (fl *Floresta) Desenhar(tela Tela) {
	for _, a := range fl.arvores {
		a.Tipo.Desenhar(tela, a.X, a.Y)
	}
}

func (fl *Floresta) Tamanho() int { return len(fl.arvores) }

// ---------- Persistencia: grava SO o contexto e a chave do flyweight ----------

type arvoreJSON struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Tipo string  `json:"tipo"` // referencia ao flyweight por nome, nao os dados dele
}

func (fl *Floresta) MarshalJSON() ([]byte, error) {
	itens := make([]arvoreJSON, len(fl.arvores))
	for i, a := range fl.arvores {
		itens[i] = arvoreJSON{a.X, a.Y, a.Tipo.nome}
	}
	return json.Marshal(itens)
}

// Carregar reconstroi a floresta pedindo os flyweights A FABRICA: o compartilhamento e restaurado.
func Carregar(dados []byte, f *FabricaTipos, catalogo map[string]TipoDef) (*Floresta, error) {
	var itens []arvoreJSON
	if err := json.Unmarshal(dados, &itens); err != nil {
		return nil, err
	}
	fl := NovaFloresta(f)
	for _, it := range itens {
		def, ok := catalogo[it.Tipo]
		if !ok {
			return nil, fmt.Errorf("tipo de arvore desconhecido: %q", it.Tipo)
		}
		fl.Plantar(it.X, it.Y, it.Tipo, def.Cor, def.Textura)
	}
	return fl, nil
}

type TipoDef struct {
	Cor     string
	Textura []byte
}
