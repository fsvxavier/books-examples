package main

import (
	"errors"
	"fmt"
	"time"
)

// temporario é o comportamento que o código de retentativa entende.
type temporario interface{ Temporario() bool }

type ErroRede struct {
	Op    string
	Causa string
}

func (e *ErroRede) Error() string    { return e.Op + ": " + e.Causa }
func (e *ErroRede) Temporario() bool { return true }

var ErrCredenciais = errors.New("credenciais inválidas")

func deveRepetir(err error) bool {
	var t temporario
	return errors.As(err, &t) && t.Temporario()
}

// Tentar repete a operação enquanto o erro for temporário, até o limite de tentativas.
// Erros definitivos interrompem na hora. O erro final guarda a causa e o número de tentativas.
func Tentar(limite int, espera time.Duration, op func(n int) error) error {
	var err error
	for n := 1; n <= limite; n++ {
		err = op(n)
		if err == nil {
			return nil
		}
		if !deveRepetir(err) {
			return fmt.Errorf("tentativa %d: erro definitivo: %w", n, err)
		}
		if n < limite {
			fmt.Printf("  tentativa %d falhou (%v), esperando %v\n", n, err, espera)
			time.Sleep(espera)
			espera *= 2
		}
	}
	return fmt.Errorf("desistindo após %d tentativas: %w", limite, err)
}

func main() {
	// 1. falha temporária que se resolve
	err := Tentar(4, time.Millisecond, func(n int) error {
		if n < 3 {
			return &ErroRede{"GET /pedidos", "conexão recusada"}
		}
		return nil
	})
	fmt.Println("resultado 1:", err)

	// 2. erro definitivo: não insiste
	err = Tentar(4, time.Millisecond, func(n int) error { return ErrCredenciais })
	fmt.Println("resultado 2:", err)
	fmt.Println(errors.Is(err, ErrCredenciais), deveRepetir(err))

	// 3. temporário que nunca se resolve
	err = Tentar(3, time.Millisecond, func(n int) error {
		return &ErroRede{"GET /pedidos", "timeout"}
	})
	fmt.Println("resultado 3:", err)
	var er *ErroRede
	fmt.Println(errors.As(err, &er), er.Op)
}
