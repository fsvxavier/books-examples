package main

import (
	"errors"
	"fmt"

	"exemplos/livro2/capitulo-2/topico-4/projeto-2/infra"
)

var ErrAmbiente = errors.New("ambiente invalido")

// Montar e a raiz de composicao: le a configuracao e devolve a FAMILIA certa.
// Recebe getenv como parametro, entao e testavel sem tocar no ambiente real.
func Montar(getenv func(string) string) (infra.Fabrica, error) {
	switch amb := getenv("APP_ENV"); amb {
	case "", "dev", "test":
		return infra.FabricaLocal{}, nil
	case "prod":
		bucket := getenv("APP_BUCKET")
		if bucket == "" {
			return nil, fmt.Errorf("%w: APP_BUCKET e obrigatorio em prod", ErrAmbiente)
		}
		return infra.FabricaNuvem{Bucket: bucket}, nil
	default:
		return nil, fmt.Errorf("%w: APP_ENV=%q (use dev, test ou prod)", ErrAmbiente, amb)
	}
}

func main() {
	envs := []map[string]string{
		{},
		{"APP_ENV": "prod", "APP_BUCKET": "relatorios"},
		{"APP_ENV": "prod"},
		{"APP_ENV": "staging"},
	}
	for _, e := range envs {
		f, err := Montar(func(k string) string { return e[k] })
		if err != nil {
			fmt.Printf("%-50s -> erro: %v\n", fmt.Sprint(e), err)
			continue
		}
		aviso, _, _ := infra.Importar(f, "x.csv", []byte("1"))
		fmt.Printf("%-50s -> %T | %s\n", fmt.Sprint(e), f, aviso)
	}
}
