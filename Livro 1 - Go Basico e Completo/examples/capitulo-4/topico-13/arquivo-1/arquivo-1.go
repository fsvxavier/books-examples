package main

import (
	"errors"
	"fmt"
	"time"
)

// Servidor reúne a configuração de um servidor HTTP hipotético.
type Servidor struct {
	endereco string
	timeout  time.Duration
	maxConex int
	logs     bool
}

// Opcao altera a configuração do Servidor e pode rejeitar valores inválidos.
type Opcao func(*Servidor) error

func ComEndereco(e string) Opcao {
	return func(s *Servidor) error {
		if e == "" {
			return errors.New("endereço vazio")
		}
		s.endereco = e
		return nil
	}
}

func ComTimeout(d time.Duration) Opcao {
	return func(s *Servidor) error {
		if d <= 0 {
			return fmt.Errorf("timeout inválido: %v", d)
		}
		s.timeout = d
		return nil
	}
}

func ComMaxConexoes(n int) Opcao {
	return func(s *Servidor) error {
		if n < 1 {
			return fmt.Errorf("máximo de conexões inválido: %d", n)
		}
		s.maxConex = n
		return nil
	}
}

func ComLogs() Opcao {
	return func(s *Servidor) error {
		s.logs = true
		return nil
	}
}

// NovoServidor parte de padrões seguros e aplica as opções na ordem recebida.
func NovoServidor(opcoes ...Opcao) (*Servidor, error) {
	s := &Servidor{
		endereco: ":8080",
		timeout:  30 * time.Second,
		maxConex: 100,
	}
	for _, aplicar := range opcoes {
		if err := aplicar(s); err != nil {
			return nil, fmt.Errorf("configurar servidor: %w", err)
		}
	}
	return s, nil
}

func main() {
	padrao, _ := NovoServidor()
	fmt.Println(padrao.endereco, padrao.timeout, padrao.maxConex, padrao.logs)

	custom, err := NovoServidor(ComEndereco(":9000"), ComTimeout(5*time.Second), ComLogs())
	fmt.Println(custom.endereco, custom.timeout, custom.maxConex, custom.logs, err)

	_, err = NovoServidor(ComMaxConexoes(0))
	fmt.Println(err)
}
