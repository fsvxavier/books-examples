package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"syscall"
)

// lerSemFechar não fecha o arquivo.
func lerSemFechar(caminho string) ([]byte, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(f)
}

func lerComDefer(caminho string) ([]byte, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// RUIM: defer dentro de laço só executa quando a função inteira retorna.
func lerTodosRuim(caminho string, n int) error {
	for range n {
		f, err := os.Open(caminho)
		if err != nil {
			return err
		}
		defer f.Close()
		io.Copy(io.Discard, f)
	}
	return nil
}

// BOM: o corpo do laço vira uma função, e o defer vale por iteração.
func lerTodosBom(caminho string, n int) error {
	for range n {
		if err := lerUm(caminho); err != nil {
			return err
		}
	}
	return nil
}

func lerUm(caminho string) error {
	f, err := os.Open(caminho)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(io.Discard, f)
	return err
}

func tentar(nome string, f func() error) {
	err := f()
	if errors.Is(err, syscall.EMFILE) {
		fmt.Printf("%-34s falhou: %v\n", nome, err)
		return
	}
	fmt.Printf("%-34s ok (err=%v)\n", nome, err)
}

func main() {
	var lim syscall.Rlimit
	syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim)
	fmt.Println("limite de descritores deste processo:", lim.Cur)

	tentar("200 leituras sem Close (sem GC)", func() error {
		for range 200 {
			if _, err := lerSemFechar("/etc/hostname"); err != nil {
				return err
			}
		}
		return nil
	})
	runtime.GC() // o finalizador de *os.File fecha o que ficou para trás
	runtime.GC()
	tentar("20 leituras sem Close (após GC)", func() error {
		for range 20 {
			if _, err := lerSemFechar("/etc/hostname"); err != nil {
				return err
			}
		}
		return nil
	})
	runtime.GC()
	tentar("200 leituras com defer", func() error {
		for range 200 {
			if _, err := lerComDefer("/etc/hostname"); err != nil {
				return err
			}
		}
		return nil
	})
	tentar("defer dentro do laço (200)", func() error { return lerTodosRuim("/etc/hostname", 200) })
	tentar("função por iteração (200)", func() error { return lerTodosBom("/etc/hostname", 200) })

	// Erro de Close/Flush ignorado: /dev/full simula disco cheio.
	fmt.Println("== escrita em disco cheio")
	f, _ := os.OpenFile("/dev/full", os.O_WRONLY, 0)
	w := bufio.NewWriter(f)
	fmt.Fprintln(w, "dado importante")
	f.Close() // sem verificar Flush: o dado se perdeu e ninguém ficou sabendo
	fmt.Println("RUIM: nenhum erro visto, dado perdido")

	f, _ = os.OpenFile("/dev/full", os.O_WRONLY, 0)
	w = bufio.NewWriter(f)
	fmt.Fprintln(w, "dado importante")
	if err := w.Flush(); err != nil {
		fmt.Println("BOM: Flush acusou:", err)
	}
	f.Close()
}
