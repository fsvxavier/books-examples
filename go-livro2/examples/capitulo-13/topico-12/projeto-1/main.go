package main

type Visitante[R any] interface {
	VisitarCidade(*Cidade) R
}

type Cidade struct{ Nome string }

// Go 1.27: um método de um tipo concreto pode ter parâmetros de tipo.
func (c *Cidade) Aceitar[R any](v Visitante[R]) R {
	return v.VisitarCidade(c)
}

// Mas o método de uma INTERFACE não pode.
type No interface {
	Aceitar[R any](v Visitante[R]) R
}

func main() {}
