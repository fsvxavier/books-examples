// Package onnx envolve o ONNX Runtime (biblioteca em C++) em uma API de Go
// pequena e segura: o resto do programa não precisa saber que há memória nativa.
package onnx

import (
	"errors"
	"fmt"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

var (
	uma     sync.Once
	errInit error
)

// Iniciar carrega a biblioteca compartilhada e cria o ambiente do runtime.
// Só a primeira chamada tem efeito; as seguintes devolvem o mesmo resultado.
func Iniciar(biblioteca string) error {
	uma.Do(func() {
		ort.SetSharedLibraryPath(biblioteca)
		errInit = ort.InitializeEnvironment()
	})
	return errInit
}

// Versao devolve a versão da biblioteca nativa carregada.
func Versao() string { return ort.GetVersion() }

// Opcoes controla as threads do runtime. Zero deixa o padrão do ONNX Runtime.
type Opcoes struct {
	ThreadsIntra int // paralelismo dentro de uma operação (uma multiplicação de matrizes)
	ThreadsInter int // paralelismo entre operações independentes do grafo
}

// Classificador roda um modelo que recebe [N, Features] floats e devolve
// [N, Classes] probabilidades.
type Classificador struct {
	sessao   *ort.DynamicAdvancedSession
	Features int
	Classes  int
}

// Abrir carrega o modelo. O custo é pago aqui, uma vez: ler o arquivo,
// otimizar o grafo e reservar memória.
func Abrir(caminho string, features, classes int, op Opcoes) (*Classificador, error) {
	opcoes, err := ort.NewSessionOptions()
	if err != nil {
		return nil, fmt.Errorf("opções da sessão: %w", err)
	}
	defer opcoes.Destroy()
	if op.ThreadsIntra > 0 {
		if err := opcoes.SetIntraOpNumThreads(op.ThreadsIntra); err != nil {
			return nil, err
		}
	}
	if op.ThreadsInter > 0 {
		if err := opcoes.SetInterOpNumThreads(op.ThreadsInter); err != nil {
			return nil, err
		}
	}
	s, err := ort.NewDynamicAdvancedSession(caminho, []string{"X"}, []string{"probabilities"}, opcoes)
	if err != nil {
		return nil, fmt.Errorf("abrir %s: %w", caminho, err)
	}
	return &Classificador{sessao: s, Features: features, Classes: classes}, nil
}

// Prever classifica um lote. Cada chamada cria os seus próprios tensores,
// então várias goroutines podem chamar Prever ao mesmo tempo no mesmo
// Classificador.
func (c *Classificador) Prever(lote [][]float32) ([][]float32, error) {
	n := len(lote)
	if n == 0 {
		return nil, nil
	}
	plano := make([]float32, 0, n*c.Features)
	for i, linha := range lote {
		if len(linha) != c.Features {
			return nil, fmt.Errorf("amostra %d tem %d valores, o modelo espera %d", i, len(linha), c.Features)
		}
		plano = append(plano, linha...)
	}
	entrada, err := ort.NewTensor(ort.NewShape(int64(n), int64(c.Features)), plano)
	if err != nil {
		return nil, err
	}
	defer entrada.Destroy() // memória nativa: o coletor de lixo do Go NÃO a libera

	saidas := []ort.Value{nil} // nil pede ao runtime que aloque a saída
	if err := c.sessao.Run([]ort.Value{entrada}, saidas); err != nil {
		return nil, fmt.Errorf("executar: %w", err)
	}
	defer saidas[0].Destroy()

	t, ok := saidas[0].(*ort.Tensor[float32])
	if !ok {
		return nil, errors.New("a saída não é um tensor de float32")
	}
	dados := t.GetData() // aponta para memória nativa: copie antes do Destroy
	res := make([][]float32, n)
	for i := range res {
		res[i] = append([]float32(nil), dados[i*c.Classes:(i+1)*c.Classes]...)
	}
	return res, nil
}

// Fechar libera a sessão e a memória nativa do modelo.
func (c *Classificador) Fechar() error { return c.sessao.Destroy() }
