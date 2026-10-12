package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Foto guarda o estado em campos não exportados (opaca para o cliente).
type Foto struct {
	titulo string
	linhas []string
}

// O encoding/json ignora campos não exportados: o resultado é "{}".
// Para persistir, a Foto precisa serializar a si mesma, com versão.
const versaoEsquema = 2

type fotoJSON struct {
	Versao int      `json:"versao"`
	Titulo string   `json:"titulo"`
	Linhas []string `json:"linhas"`
}

func (f Foto) MarshalJSON() ([]byte, error) {
	return json.Marshal(fotoJSON{versaoEsquema, f.titulo, f.linhas})
}

var ErrVersao = errors.New("foto de versão desconhecida")

func (f *Foto) UnmarshalJSON(b []byte) error {
	var j fotoJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	switch j.Versao {
	case 2:
	case 1:
		// versão 1 não tinha título: migra com um valor padrão
		j.Titulo = "(sem título)"
	default:
		return fmt.Errorf("%w: %d", ErrVersao, j.Versao)
	}
	f.titulo, f.linhas = j.Titulo, j.Linhas
	return nil
}

// Sem MarshalJSON, a mesma estrutura serializa vazia.
type FotoMuda struct {
	titulo string
	linhas []string
}

// gravarAtomico escreve em arquivo temporário e renomeia.
func gravarAtomico(caminho string, dados []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(caminho), "foto-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(dados); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), caminho)
}

func main() {
	muda, _ := json.Marshal(FotoMuda{"x", []string{"a"}})
	fmt.Println("sem MarshalJSON:", string(muda))

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(FotoMuda{"x", []string{"a"}})
	fmt.Println("gob:", err)

	f := Foto{"relatório", []string{"linha 1", "linha 2"}}
	dados, _ := json.Marshal(f)
	fmt.Println("com MarshalJSON:", string(dados))

	dir, _ := os.MkdirTemp("", "cap34")
	caminho := filepath.Join(dir, "checkpoint.json")
	if err := gravarAtomico(caminho, dados); err != nil {
		fmt.Println(err)
		return
	}
	lido, _ := os.ReadFile(caminho)
	var g Foto
	fmt.Println("recarregou:", json.Unmarshal(lido, &g), g.titulo, g.linhas)

	var v1 Foto
	fmt.Println("v1:", json.Unmarshal([]byte(`{"versao":1,"linhas":["velha"]}`), &v1), v1.titulo, v1.linhas)

	var v9 Foto
	err = json.Unmarshal([]byte(`{"versao":9}`), &v9)
	fmt.Println("v9:", err, errors.Is(err, ErrVersao))
}
