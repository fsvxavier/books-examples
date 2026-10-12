package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

const (
	nArvores = 100_000
	tamTex   = 1024 // bytes de "textura" por tipo de arvore
)

// ---------- SEM Flyweight: cada arvore carrega o seu proprio tipo ----------

type ArvoreCheia struct {
	X, Y    float32
	Nome    string
	Cor     string
	Textura []byte // copia propria
}

// ---------- COM Flyweight ----------

// TipoArvore guarda o estado INTRINSECO: igual para todas as arvores do mesmo tipo e imutavel.
type TipoArvore struct {
	nome    string
	cor     string
	textura []byte
}

func (t *TipoArvore) Nome() string { return t.nome }

// Desenhar recebe o estado EXTRINSECO (x, y) como parametro.
func (t *TipoArvore) Desenhar(x, y float32) int {
	return len(t.textura) + int(x) + int(y) // finge desenhar
}

// Arvore e o CONTEXTO: so o que muda de uma arvore para outra, mais a referencia ao flyweight.
type Arvore struct {
	X, Y float32
	Tipo *TipoArvore
}

type Floresta struct{ arvores []Arvore }

func novaTextura(semente byte) []byte {
	b := make([]byte, tamTex)
	for i := range b {
		b[i] = semente + byte(i)
	}
	return b
}

// medir devolve quantos bytes de heap o resultado de f ocupa (depois de um GC).
func medir(f func() any) (any, uint64) {
	var antes, depois runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&antes)
	v := f()
	runtime.GC()
	runtime.ReadMemStats(&depois)
	return v, depois.HeapAlloc - antes.HeapAlloc
}

func main() {
	tipos := []struct{ nome, cor string }{{"carvalho", "verde-escuro"}, {"pinheiro", "verde"}, {"bétula", "branco"}}

	v1, bytesCheia := medir(func() any {
		fl := make([]ArvoreCheia, 0, nArvores)
		for i := range nArvores {
			t := tipos[i%3]
			fl = append(fl, ArvoreCheia{X: float32(i % 1000), Y: float32(i / 1000), Nome: t.nome, Cor: t.cor, Textura: novaTextura(byte(i % 3))})
		}
		return fl
	})

	v2, bytesFly := medir(func() any {
		// A "fabrica" aqui e um slice com os 3 tipos (a secao 30.x mostra uma fabrica com mapa).
		pool := make([]*TipoArvore, 3)
		for i, t := range tipos {
			pool[i] = &TipoArvore{nome: t.nome, cor: t.cor, textura: novaTextura(byte(i))}
		}
		f := &Floresta{arvores: make([]Arvore, 0, nArvores)}
		for i := range nArvores {
			f.arvores = append(f.arvores, Arvore{X: float32(i % 1000), Y: float32(i / 1000), Tipo: pool[i%3]})
		}
		return f
	})

	fmt.Printf("%d arvores, 3 tipos, textura de %d bytes por tipo\n", nArvores, tamTex)
	fmt.Printf("sem Flyweight: %8.2f MiB  (%d bytes por arvore)\n", float64(bytesCheia)/(1<<20), bytesCheia/nArvores)
	fmt.Printf("com Flyweight: %8.2f MiB  (%d bytes por arvore)\n", float64(bytesFly)/(1<<20), bytesFly/nArvores)
	fmt.Printf("reducao: %.0f vezes\n", float64(bytesCheia)/float64(bytesFly))

	// O comportamento e o mesmo: o resultado do desenho confere nas duas formas.
	var somaCheia, somaFly int
	for _, a := range v1.([]ArvoreCheia) {
		somaCheia += len(a.Textura) + int(a.X) + int(a.Y)
	}
	for _, a := range v2.(*Floresta).arvores {
		somaFly += a.Tipo.Desenhar(a.X, a.Y)
	}
	fmt.Println("mesmo resultado de desenho:", somaCheia == somaFly)
	fmt.Println("unsafe.Sizeof(Arvore) =", unsafe.Sizeof(Arvore{}), "bytes | unsafe.Sizeof(ArvoreCheia) =", unsafe.Sizeof(ArvoreCheia{}), "bytes (sem contar a textura)")
}
