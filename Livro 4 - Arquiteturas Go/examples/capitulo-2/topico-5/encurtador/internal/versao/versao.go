// Package versao identifica o binário em execução.
package versao

import (
	"fmt"
	"runtime/debug"
)

// Versao é preenchida na compilação:
//
//	go build -ldflags "-X exemplos/livro4/capitulo-2/topico-5/encurtador/internal/versao.versao=1.2.3"
//
// É uma das raras variáveis de pacote aceitáveis: vale um valor só, definido
// antes de o programa começar e nunca alterado depois.
var Versao = "dev"

// Texto descreve a versão, o Go usado e, se houver, a revisão do controle de versão.
func Texto() string {
	texto := "versão " + Versao
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return texto
	}
	texto += fmt.Sprintf(" (%s", info.GoVersion)
	modificado := false
	for _, c := range info.Settings {
		switch c.Key {
		case "vcs.revision":
			if len(c.Value) >= 7 {
				texto += ", rev " + c.Value[:7]
			}
		case "vcs.modified":
			modificado = c.Value == "true"
		}
	}
	if modificado {
		texto += ", com alterações não commitadas"
	}
	return texto + ")"
}
