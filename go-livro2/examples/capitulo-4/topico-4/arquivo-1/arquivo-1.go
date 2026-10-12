package main

import (
	"fmt"
	"sync"
)

var tentativas int

func conectar() (string, error) {
	tentativas++
	if tentativas < 3 {
		return "", fmt.Errorf("banco indisponivel (tentativa %d)", tentativas)
	}
	return "conexao-ok", nil
}

// 1) sync.OnceValues guarda o erro para sempre.
func umaVez() {
	obter := sync.OnceValues(conectar)
	for i := 1; i <= 3; i++ {
		v, err := obter()
		fmt.Printf("OnceValues chamada %d: valor=%q erro=%v\n", i, v, err)
	}
	fmt.Println("conectar executou", tentativas, "vez")
}

// 2) Once.Do com panic: a funcao conta como executada.
func comPanic() {
	var o sync.Once
	chamadas := 0
	f := func() {
		chamadas++
		panic("falhou na inicializacao")
	}
	for i := 1; i <= 3; i++ {
		func() {
			defer func() {
				fmt.Printf("Once.Do chamada %d: recover=%v\n", i, recover())
			}()
			o.Do(f)
		}()
	}
	fmt.Println("f executou", chamadas, "vez")
}

// 3) OnceValue com panic: o mesmo panic e repetido em toda chamada.
func onceValueComPanic() {
	chamadas := 0
	obter := sync.OnceValue(func() int {
		chamadas++
		panic("config invalida")
	})
	for i := 1; i <= 3; i++ {
		func() {
			defer func() {
				fmt.Printf("OnceValue chamada %d: recover=%v\n", i, recover())
			}()
			obter()
		}()
	}
	fmt.Println("funcao executou", chamadas, "vez")
}

// 4) Singleton que tenta de novo: mutex em vez de Once.
type Preguicoso[T any] struct {
	mu     sync.Mutex
	valor  T
	pronto bool
	novo   func() (T, error)
}

func (p *Preguicoso[T]) Obter() (T, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.pronto {
		return p.valor, nil
	}
	v, err := p.novo()
	if err != nil {
		var zero T
		return zero, err
	}
	p.valor, p.pronto = v, true
	return v, nil
}

func comRetentativa() {
	tentativas = 0
	p := &Preguicoso[string]{novo: conectar}
	for i := 1; i <= 4; i++ {
		v, err := p.Obter()
		fmt.Printf("Preguicoso chamada %d: valor=%q erro=%v\n", i, v, err)
	}
	fmt.Println("conectar executou", tentativas, "vezes")
}

func main() {
	umaVez()
	fmt.Println()
	comPanic()
	fmt.Println()
	onceValueComPanic()
	fmt.Println()
	comRetentativa()
}
