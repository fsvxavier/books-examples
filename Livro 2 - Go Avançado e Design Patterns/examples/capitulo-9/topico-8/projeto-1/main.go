package main

import "fmt"

func main() {
	ed := NovoEditor("Ola")
	h := &Historico{}

	mostrar := func(rotulo string) {
		f, d := h.Tamanho()
		fmt.Printf("%-34s texto=%-22q feitos=%d desfeitos=%d\n", rotulo, ed.Texto(), f, d)
	}
	mostrar("inicio")

	h.Executar(&InserirTexto{ed, 3, ", mundo"})
	mostrar("inserir ', mundo'")
	h.Executar(&InserirTexto{ed, 10, "!"})
	mostrar("inserir '!'")
	h.Executar(&Maiusculas{Ed: ed})
	mostrar("maiusculas (por backup)")
	h.Executar(&ApagarTrecho{Ed: ed, Pos: 0, N: 4})
	mostrar("apagar 4 no inicio")

	h.Desfazer()
	mostrar("desfazer")
	h.Desfazer()
	mostrar("desfazer (maiusculas voltam)")
	h.Refazer()
	mostrar("refazer")

	h.Executar(&InserirTexto{ed, 0, ">> "})
	mostrar("comando NOVO apos desfazer")
	fmt.Println("erro ao refazer depois de um comando novo:", h.Refazer())

	err := h.Executar(&ApagarTrecho{Ed: ed, Pos: 50, N: 3})
	f, _ := h.Tamanho()
	fmt.Printf("comando que falha: %v | feitos continua %d\n", err, f)

	for h.Desfazer() == nil {
	}
	mostrar("desfeito tudo")
}
