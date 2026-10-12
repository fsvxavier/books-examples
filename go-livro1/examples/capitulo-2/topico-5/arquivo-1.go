package main

import "fmt"

// DiaSemana é um tipo próprio, o que impede misturar dias com inteiros quaisquer.
type DiaSemana int

const (
	Domingo DiaSemana = iota
	Segunda
	Terca
	Quarta
	Quinta
	Sexta
	Sabado
)

var nomesDias = [...]string{"domingo", "segunda", "terça", "quarta", "quinta", "sexta", "sábado"}

// String faz DiaSemana implementar fmt.Stringer, e o Println passa a usá-lo.
func (d DiaSemana) String() string {
	if d < Domingo || d > Sabado {
		return fmt.Sprintf("DiaSemana(%d)", int(d))
	}
	return nomesDias[d]
}

func main() {
	fmt.Println(Domingo, Quarta, Sabado, DiaSemana(9))
	fmt.Printf("%d %v %s\n", Terca, Terca, Terca)
}
