package main

import (
	"bytes"
	"fmt"
)

func main() {
	contas := NovasContas()
	diario := &Diario{}
	proc := NovoProcessador(contas, diario)
	fila := NovaFila(proc, 16)

	enviar := func(id string, c Comando) {
		e, _ := Empacotar(id, c)
		fila.Enviar(e)
	}
	enviar("c1", Depositar{"ana", 10000})
	enviar("c2", Depositar{"bia", 2500})
	enviar("c3", Sacar{"ana", 3000})
	enviar("c4", Sacar{"bia", 9999}) // vai falhar: saldo insuficiente
	enviar("c3", Sacar{"ana", 3000}) // repeticao do c3: deve ser ignorada
	fila.Encerrar()

	fmt.Println("saldos:", contas.Snapshot())
	fmt.Println("comandos que falharam:", proc.Falhas)

	fmt.Println("--- o diario (um comando por linha, em JSON):")
	fmt.Print(string(diario.Bytes()))

	// Reconstroi o estado a partir do diario, em contas novas.
	refeitas, err := Reproduzir(bytes.NewReader(diario.Bytes()))
	fmt.Println("--- estado reconstruido pelo diario:", refeitas.Snapshot(), "erro:", err)

	// Um tipo desconhecido (por exemplo, de uma versao futura) e rejeitado com clareza.
	_, err = Reproduzir(bytes.NewReader([]byte(`{"id":"x","tipo":"transferir","dados":{}}` + "\n")))
	fmt.Println("tipo desconhecido:", err)
}
