// Teste mínimo do ONNX Runtime a partir do Go: carrega um modelo de uma única
// operação (Add) e soma dois vetores. Precisa de cgo e da biblioteca compartilhada.
//
// uso: go run . CAMINHO/libonnxruntime.so soma.onnx
package main

import (
	"fmt"
	"os"

	ort "github.com/yalue/onnxruntime_go"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "uso: go run . CAMINHO/libonnxruntime.so soma.onnx")
		os.Exit(2)
	}
	ort.SetSharedLibraryPath(os.Args[1])
	if err := ort.InitializeEnvironment(); err != nil {
		fmt.Fprintln(os.Stderr, "iniciar:", err)
		os.Exit(1)
	}
	defer ort.DestroyEnvironment()

	x, _ := ort.NewTensor(ort.NewShape(3), []float32{1, 2, 3})
	y, _ := ort.NewTensor(ort.NewShape(3), []float32{10, 20, 30})
	z, _ := ort.NewEmptyTensor[float32](ort.NewShape(3))
	defer x.Destroy()
	defer y.Destroy()
	defer z.Destroy()

	sessao, err := ort.NewAdvancedSession(os.Args[2], []string{"x", "y"}, []string{"z"},
		[]ort.Value{x, y}, []ort.Value{z}, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sessão:", err)
		os.Exit(1)
	}
	defer sessao.Destroy()

	if err := sessao.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "run:", err)
		os.Exit(1)
	}
	fmt.Println("onnxruntime", ort.GetVersion(), "->", z.GetData())
}
