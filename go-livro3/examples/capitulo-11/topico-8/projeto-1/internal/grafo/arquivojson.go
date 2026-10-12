package grafo

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// ArquivoJSON guarda o ponto de execução em um arquivo JSON.
// A gravação é atômica: escreve em um arquivo temporário e renomeia.
type ArquivoJSON[S any] struct {
	Caminho string
}

type registro[S any] struct {
	Proximo string `json:"proximo"`
	Estado  S      `json:"estado"`
}

// Guardar implementa Ponto.
func (a ArquivoJSON[S]) Guardar(proximo string, s S) error {
	dados, err := json.Marshal(registro[S]{Proximo: proximo, Estado: s})
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(a.Caminho), ".ponto-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(dados); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), a.Caminho)
}

// Carregar implementa Ponto.
func (a ArquivoJSON[S]) Carregar() (string, S, bool, error) {
	var r registro[S]
	dados, err := os.ReadFile(a.Caminho)
	if errors.Is(err, fs.ErrNotExist) {
		return "", r.Estado, false, nil
	}
	if err != nil {
		return "", r.Estado, false, err
	}
	if err := json.Unmarshal(dados, &r); err != nil {
		return "", r.Estado, false, err
	}
	return r.Proximo, r.Estado, true, nil
}
