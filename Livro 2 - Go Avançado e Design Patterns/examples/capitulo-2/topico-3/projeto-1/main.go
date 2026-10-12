package main

import (
	"fmt"

	"exemplos/livro2/capitulo-2/topico-3/projeto-1/infra"
)

func escolher(ambiente string) infra.Fabrica {
	if ambiente == "producao" {
		return infra.FabricaNuvem{Bucket: "relatorios"}
	}
	return infra.FabricaLocal{}
}

func main() {
	for _, amb := range []string{"desenvolvimento", "producao"} {
		aviso, lido, err := infra.Importar(escolher(amb), "vendas.csv", []byte("a;b"))
		fmt.Printf("%-16s aviso=%q lido=%q err=%v\n", amb, aviso, lido, err)
	}
	fmt.Println("misturando familias:", infra.ImportarMisturado("vendas.csv", []byte("a;b")))
}
