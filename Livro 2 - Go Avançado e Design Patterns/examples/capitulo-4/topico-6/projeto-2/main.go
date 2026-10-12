package main

import (
	"fmt"
	"os"
	"time"
)

func abrirPool() string {
	time.Sleep(300 * time.Millisecond) // handshake, autenticacao, aquecimento
	return "pool"
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "versao" {
		fmt.Println("v1.0.0")
		return
	}
	fmt.Println("consulta com", pool())
}
