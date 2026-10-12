package db

import (
	"fmt"
	"time"

	"github.com/fsvxavier/books-examples/go-livro-2/capitulo-4/topico-6/projeto-1/log"
)

var Pool = abrirPool()

func abrirPool() string {
	fmt.Println("db: abrindo pool (caro, nivel de log =", log.Nivel+")")
	time.Sleep(300 * time.Millisecond)
	return "pool"
}

func init() {
	fmt.Println("db: init()")
}
