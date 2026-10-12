// Package consulta monta SELECTs parametrizados com um builder fluente.
// Os VALORES nunca entram no texto SQL: vao em Args, com placeholders $1, $2...
// Os IDENTIFICADORES (colunas de ordenacao) so entram se estiverem numa lista permitida.
package consulta

import (
	"errors"
	"fmt"
	"strings"
)

type Consulta struct {
	SQL  string
	Args []any
}

type Builder struct {
	tabela     string
	colunas    []string
	filtros    []string
	args       []any
	ordem      string
	limite     int
	deslocar   int
	permitidas map[string]bool
	erros      []error
}

// De inicia o builder. permitidas e a lista de colunas aceitas em ORDER BY.
func De(tabela string, permitidas ...string) *Builder {
	b := &Builder{tabela: tabela, permitidas: map[string]bool{}, limite: -1}
	for _, c := range permitidas {
		b.permitidas[c] = true
	}
	return b
}

func (b *Builder) Colunas(cs ...string) *Builder { b.colunas = append(b.colunas, cs...); return b }

// Onde acrescenta uma condicao com UM placeholder. O texto vem do programador; o valor, do usuario.
func (b *Builder) Onde(condicao string, valor any) *Builder {
	if strings.Count(condicao, "?") != 1 {
		b.erros = append(b.erros, fmt.Errorf("condicao %q deve ter exatamente um '?'", condicao))
		return b
	}
	b.args = append(b.args, valor)
	b.filtros = append(b.filtros, strings.Replace(condicao, "?", fmt.Sprintf("$%d", len(b.args)), 1))
	return b
}

// OndeSe so aplica o filtro quando a condicao Go e verdadeira (filtros opcionais).
func (b *Builder) OndeSe(aplicar bool, condicao string, valor any) *Builder {
	if aplicar {
		return b.Onde(condicao, valor)
	}
	return b
}

func (b *Builder) OrdenarPor(coluna string, decrescente bool) *Builder {
	if !b.permitidas[coluna] {
		b.erros = append(b.erros, fmt.Errorf("ordenar por %q nao permitido", coluna))
		return b
	}
	b.ordem = coluna
	if decrescente {
		b.ordem += " DESC"
	}
	return b
}

func (b *Builder) Limite(n int) *Builder {
	if n < 1 || n > 1000 {
		b.erros = append(b.erros, fmt.Errorf("limite %d fora de 1..1000", n))
		return b
	}
	b.limite = n
	return b
}

func (b *Builder) Deslocar(n int) *Builder {
	if n < 0 {
		b.erros = append(b.erros, fmt.Errorf("deslocamento %d negativo", n))
		return b
	}
	b.deslocar = n
	return b
}

// Construir valida tudo de uma vez e devolve o resultado IMUTAVEL.
func (b *Builder) Construir() (Consulta, error) {
	if b.tabela == "" {
		b.erros = append(b.erros, errors.New("tabela obrigatoria"))
	}
	if b.deslocar > 0 && b.limite < 0 {
		b.erros = append(b.erros, errors.New("deslocamento exige limite"))
	}
	if len(b.erros) > 0 {
		return Consulta{}, errors.Join(b.erros...)
	}
	cols := "*"
	if len(b.colunas) > 0 {
		cols = strings.Join(b.colunas, ", ")
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "SELECT %s FROM %s", cols, b.tabela)
	if len(b.filtros) > 0 {
		sb.WriteString(" WHERE " + strings.Join(b.filtros, " AND "))
	}
	if b.ordem != "" {
		sb.WriteString(" ORDER BY " + b.ordem)
	}
	args := append([]any(nil), b.args...) // copia: o builder pode continuar sendo usado
	if b.limite > 0 {
		args = append(args, b.limite)
		fmt.Fprintf(&sb, " LIMIT $%d", len(args))
		if b.deslocar > 0 {
			args = append(args, b.deslocar)
			fmt.Fprintf(&sb, " OFFSET $%d", len(args))
		}
	}
	return Consulta{SQL: sb.String(), Args: args}, nil
}
