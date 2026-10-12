package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	ctx := context.Background()
	cadeia, _ := Fabrica("robo", "atendente", "especialista")

	chamados := []Chamado{
		{"Esqueci minha senha", 1, ""},
		{"Impressora nao liga", 2, ""},
		{"Banco de dados corrompido", 4, ""},
		{"Servidor em chamas", 5, ""},
	}
	for _, c := range chamados {
		err := cadeia.Tratar(ctx, &c)
		fmt.Printf("%-28s gravidade %d -> tratado por %-12q err=%v\n", c.Assunto, c.Gravidade, c.TratadoPor, errors.Is(err, ErrNaoTratado))
	}

	// A cadeia e DADO: muda em tempo de execucao sem mexer nos tratadores.
	fmt.Println("--- acrescentando a coordenacao ao fim da cadeia")
	nova := cadeia.Inserir(len(cadeia), PorGravidade("coordenacao", 5))
	c := Chamado{"Servidor em chamas", 5, ""}
	err := nova.Tratar(ctx, &c)
	fmt.Printf("%-28s gravidade %d -> tratado por %q err=%v\n", c.Assunto, c.Gravidade, c.TratadoPor, err)
	fmt.Println("a cadeia original continua com", len(cadeia), "elos; a nova tem", len(nova))

	fmt.Println("--- ordem importa: especialista ANTES do robo captura tudo de gravidade <= 4")
	invertida, _ := Fabrica("especialista", "robo")
	c = Chamado{"Esqueci minha senha", 1, ""}
	invertida.Tratar(ctx, &c)
	fmt.Printf("%-28s tratado por %q (o robo nunca foi consultado)\n", c.Assunto, c.TratadoPor)

	_, err = Fabrica("robo", "gerente")
	fmt.Println("nome desconhecido:", err)
}
