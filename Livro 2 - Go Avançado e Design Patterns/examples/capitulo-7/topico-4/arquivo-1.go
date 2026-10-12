package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

// lerManual faz, passo a passo, o que os.ReadFile faz por voce.
func lerManual(nome string) ([]byte, error) {
	f, err := os.Open(nome)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tamanho int
	if info, err := f.Stat(); err == nil {
		tamanho = int(info.Size())
	}
	dados := make([]byte, 0, tamanho+1) // +1 para detectar o EOF sem realocar
	for {
		n, err := f.Read(dados[len(dados):cap(dados)])
		dados = dados[:len(dados)+n]
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return dados, err
		}
		if len(dados) >= cap(dados) {
			dados = append(dados, 0)[:len(dados)]
		}
	}
}

func main() {
	tmp, _ := os.CreateTemp("", "fachada-*.txt")
	defer os.Remove(tmp.Name())
	tmp.WriteString("conteudo do arquivo\n")
	tmp.Close()

	a, _ := lerManual(tmp.Name())
	b, _ := os.ReadFile(tmp.Name()) // a fachada: uma linha
	fmt.Printf("manual=%q fachada=%q iguais=%v\n", a, b, string(a) == string(b))

	_, err := os.ReadFile("/nao/existe.txt")
	fmt.Println("erro da fachada ainda e tipado:", errors.Is(err, os.ErrNotExist))

	// http.Get e outra fachada: esconde Client, Request, Transport, conexao...
	lento := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		fmt.Fprint(w, "ok")
	}))
	defer lento.Close()

	inicio := time.Now()
	resp, err := http.Get(lento.URL)
	if err == nil {
		resp.Body.Close()
	}
	fmt.Printf("http.Get esperou o servidor lento: %v (erro=%v)\n", time.Since(inicio).Round(100*time.Millisecond), err)

	// A fachada simplifica o caso comum; para controlar o prazo, desca ao subsistema.
	cliente := &http.Client{Timeout: 50 * time.Millisecond}
	inicio = time.Now()
	_, err = cliente.Get(lento.URL)
	fmt.Printf("Client com Timeout desistiu em ~%v: %v\n", time.Since(inicio).Round(50*time.Millisecond), err != nil)
}
