package longo

import (
	"regexp"
	"strings"
)

// Fato é uma informação durável sobre o usuário, extraída de uma fala.
type Fato struct {
	Tipo        string
	Conteudo    string
	Importancia float32
}

// padroes são as regras do extrator: uma expressão regular que captura o valor,
// o tipo do fato, o modelo canônico da frase e a importância.
var padroes = []struct {
	re          *regexp.Regexp
	tipo        string
	modelo      string
	importancia float32
}{
	{regexp.MustCompile(`(?i)\bmeu nome é ([\p{L}]+)`), "nome", "O usuário se chama %s", 0.9},
	{regexp.MustCompile(`(?i)\bmoro em ([\p{L}][\p{L} ]*[\p{L}])`), "local", "O usuário mora em %s", 0.7},
	{regexp.MustCompile(`(?i)\btrabalho com ([\p{L}][\p{L} ]*[\p{L}])`), "trabalho", "O usuário trabalha com %s", 0.7},
	{regexp.MustCompile(`(?i)\bprefiro ([\p{L}][\p{L} ]*[\p{L}])`), "preferencia", "O usuário prefere %s", 0.6},
	{regexp.MustCompile(`(?i)\bnão gosto de ([\p{L}][\p{L} ]*[\p{L}])`), "aversao", "O usuário não gosta de %s", 0.6},
}

// Extrair procura fatos duráveis em uma fala do usuário.
func Extrair(fala string) []Fato {
	var fatos []Fato
	for _, p := range padroes {
		for _, m := range p.re.FindAllStringSubmatch(fala, -1) {
			fatos = append(fatos, Fato{
				Tipo:        p.tipo,
				Conteudo:    strings.Replace(p.modelo, "%s", strings.TrimSpace(m[1]), 1),
				Importancia: p.importancia,
			})
		}
	}
	return fatos
}
