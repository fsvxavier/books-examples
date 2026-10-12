// Leitura ingênua: carrega o arquivo inteiro e decodifica linha a linha.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var docs []map[string]any
	for _, linha := range bytes.Split(b, []byte("\n")) {
		var d map[string]any
		if json.Unmarshal(linha, &d) == nil {
			docs = append(docs, d)
		}
	}
	fmt.Println("documentos:", len(docs))
}
