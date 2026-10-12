// Command gerar lê um arquivo Go e escreve, para cada struct com tags "validate",
// um método Validar sem reflexão. É a metaprogramação feita em tempo de compilação.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type regra struct {
	Nome string
	Arg  string
}

type campo struct {
	Nome   string
	Tipo   string // "string" ou "int"
	Regras []regra
}

type estrutura struct {
	Nome   string
	Campos []campo
}

func main() {
	entrada := flag.String("arquivo", "", "arquivo Go com as structs")
	saida := flag.String("saida", "", "arquivo a escrever")
	flag.Parse()
	if *entrada == "" || *saida == "" {
		fmt.Fprintln(os.Stderr, "uso: gerar -arquivo modelo.go -saida validar_gen.go")
		os.Exit(2)
	}
	if err := executar(*entrada, *saida); err != nil {
		fmt.Fprintln(os.Stderr, "gerar:", err)
		os.Exit(1)
	}
}

func executar(entrada, saida string) error {
	fset := token.NewFileSet()
	arq, err := parser.ParseFile(fset, entrada, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	var estruturas []estrutura
	for _, d := range arq.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts := spec.(*ast.TypeSpec)
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			e, err := ler(ts.Name.Name, st)
			if err != nil {
				return fmt.Errorf("%s: %w", fset.Position(ts.Pos()), err)
			}
			if len(e.Campos) > 0 {
				estruturas = append(estruturas, e)
			}
		}
	}
	var buf bytes.Buffer
	if err := modelo.Execute(&buf, struct {
		Pacote     string
		Estruturas []estrutura
	}{arq.Name.Name, estruturas}); err != nil {
		return err
	}
	codigo, err := format.Source(buf.Bytes()) // o arquivo gerado já sai formatado
	if err != nil {
		return fmt.Errorf("o código gerado não compila: %w\n%s", err, buf.Bytes())
	}
	return os.WriteFile(saida, codigo, 0o644)
}

func ler(nome string, st *ast.StructType) (estrutura, error) {
	e := estrutura{Nome: nome}
	for _, f := range st.Fields.List {
		if f.Tag == nil || len(f.Names) != 1 {
			continue
		}
		valor, _ := strconv.Unquote(f.Tag.Value)
		tag, ok := reflect.StructTag(valor).Lookup("validate")
		if !ok || tag == "-" {
			continue
		}
		id, ok := f.Type.(*ast.Ident)
		if !ok || (id.Name != "string" && id.Name != "int") {
			return e, fmt.Errorf("o campo %s tem um tipo que o gerador não suporta (só string e int)", f.Names[0].Name)
		}
		c := campo{Nome: f.Names[0].Name, Tipo: id.Name}
		for _, parte := range strings.Split(tag, ",") {
			n, arg, _ := strings.Cut(strings.TrimSpace(parte), "=")
			switch n {
			case "obrigatorio", "email":
			case "min", "max":
				if _, err := strconv.Atoi(arg); err != nil {
					return e, fmt.Errorf("campo %s: %s=%q não é um inteiro", c.Nome, n, arg)
				}
			default:
				return e, fmt.Errorf("campo %s: regra desconhecida %q", c.Nome, n)
			}
			if n == "email" && id.Name != "string" {
				return e, fmt.Errorf("campo %s: email só vale para string", c.Nome)
			}
			c.Regras = append(c.Regras, regra{n, arg})
		}
		e.Campos = append(e.Campos, c)
	}
	return e, nil
}

var modelo = template.Must(template.New("validar").Parse(`// Código gerado por gerar. NÃO EDITE.

package {{.Pacote}}

import (
	"net/mail"
	"unicode/utf8"

	"exemplo.com/reflexao/e06"
)

var _ = mail.ParseAddress
var _ = utf8.RuneCountInString
{{range $e := .Estruturas}}
// Validar verifica as tags "validate" de {{$e.Nome}}, sem reflexão.
func (x *{{$e.Nome}}) Validar() e06.Falhas {
	var fs e06.Falhas
	{{- range $c := $e.Campos}}
	{{- range $r := $c.Regras}}
	{{- if eq $r.Nome "obrigatorio"}}
	{{- if eq $c.Tipo "string"}}
	if x.{{$c.Nome}} == "" {
		fs = append(fs, e06.Falha{Campo: "{{$c.Nome}}", Regra: "obrigatorio", Mensagem: "é obrigatório"})
	}
	{{- else}}
	if x.{{$c.Nome}} == 0 {
		fs = append(fs, e06.Falha{Campo: "{{$c.Nome}}", Regra: "obrigatorio", Mensagem: "é obrigatório"})
	}
	{{- end}}
	{{- else if eq $r.Nome "min"}}
	{{- if eq $c.Tipo "string"}}
	if utf8.RuneCountInString(x.{{$c.Nome}}) < {{$r.Arg}} {
		fs = append(fs, e06.Falha{Campo: "{{$c.Nome}}", Regra: "min", Mensagem: "abaixo do mínimo de {{$r.Arg}}"})
	}
	{{- else}}
	if x.{{$c.Nome}} < {{$r.Arg}} {
		fs = append(fs, e06.Falha{Campo: "{{$c.Nome}}", Regra: "min", Mensagem: "abaixo do mínimo de {{$r.Arg}}"})
	}
	{{- end}}
	{{- else if eq $r.Nome "max"}}
	{{- if eq $c.Tipo "string"}}
	if utf8.RuneCountInString(x.{{$c.Nome}}) > {{$r.Arg}} {
		fs = append(fs, e06.Falha{Campo: "{{$c.Nome}}", Regra: "max", Mensagem: "acima do máximo de {{$r.Arg}}"})
	}
	{{- else}}
	if x.{{$c.Nome}} > {{$r.Arg}} {
		fs = append(fs, e06.Falha{Campo: "{{$c.Nome}}", Regra: "max", Mensagem: "acima do máximo de {{$r.Arg}}"})
	}
	{{- end}}
	{{- else if eq $r.Nome "email"}}
	if x.{{$c.Nome}} != "" {
		if a, err := mail.ParseAddress(x.{{$c.Nome}}); err != nil || a.Address != x.{{$c.Nome}} {
			fs = append(fs, e06.Falha{Campo: "{{$c.Nome}}", Regra: "email", Mensagem: "não é um e-mail válido"})
		}
	}
	{{- end}}
	{{- end}}
	{{- end}}
	return fs
}
{{end}}`))
