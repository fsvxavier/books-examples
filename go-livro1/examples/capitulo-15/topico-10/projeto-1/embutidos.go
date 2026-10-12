package arq

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"text/template"
)

//go:embed assets/email.tmpl
var modeloEmail string

//go:embed assets/sql/*.sql
var migracoes embed.FS

// Migracoes lista os arquivos de migração embutidos no binário, em ordem.
func Migracoes() ([]string, error) {
	var nomes []string
	err := fs.WalkDir(migracoes, "assets/sql", func(caminho string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			nomes = append(nomes, caminho)
		}
		return nil
	})
	return nomes, err
}

// Email renderiza o modelo embutido.
func Email(nome string, id int) (string, error) {
	t, err := template.New("email").Parse(modeloEmail)
	if err != nil {
		return "", fmt.Errorf("modelo inválido: %w", err)
	}
	var sb strings.Builder
	err = t.Execute(&sb, struct {
		Nome string
		ID   int
	}{nome, id})
	return strings.TrimSpace(sb.String()), err
}
