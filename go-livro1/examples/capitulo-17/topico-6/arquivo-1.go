package v1

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Contador struct {
	mu sync.Mutex
	n  int
}

// copylocks: o receptor por valor copia o Mutex.
func (c Contador) Valor() int { return c.n }

func Imprime(nome string, idade int) {
	fmt.Printf("nome=%s idade=%s\n", nome, idade) // printf: %s com int
	fmt.Printf("sem argumento: %d\n")             // printf: faltou o argumento
}

func Espera() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second) // lostcancel
	_ = ctx
}

func Pega(url string) error {
	resp, err := http.Get(url)
	defer resp.Body.Close() // httpresponse: usou resp antes de checar err
	if err != nil {
		return err
	}
	return nil
}

func Datas(t time.Time) string {
	return t.Format("2006-02-01") // timeformat: dia e mês trocados (o certo é 2006-01-02)
}

func Registra(l *slog.Logger) {
	l.Info("pedido criado", "id", 7, "valor") // slog: chave sem valor
}

type Cfg struct {
	Nome string `json:"nome" xml`
}

func Erros(err error) {
	var alvo error
	errors.As(err, alvo) // errorsas: o segundo argumento deve ser ponteiro
}

func Inalcancavel() {
	return
	fmt.Println("nunca roda") // unreachable
}

func Grupo() {
	var wg sync.WaitGroup
	go func() {
		wg.Add(1) // waitgroup: Add dentro da goroutine
		defer wg.Done()
	}()
	wg.Wait()
}
