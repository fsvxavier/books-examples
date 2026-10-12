// Package checkout monta o fechamento de compra de acordo com a regiao.
// Abstract Factory: cada regiao e uma FAMILIA de regras que precisam combinar entre si.
// Factory Method: o meio de pagamento e escolhido por nome dentro da regiao.
package checkout

import (
	"fmt"
	"regexp"
	"strings"
)

// ---- Produtos da familia ----

type ValidadorDocumento interface{ Validar(doc string) error }

type Formatador interface{ Moeda(centavos int64) string }

type CalculadoraImposto interface{ Imposto(centavos int64) int64 }

type MeioPagamento interface{ Autorizar(centavos int64) string }

// ---- A fabrica abstrata ----

type Regiao interface {
	Nome() string
	Documento() ValidadorDocumento
	Formato() Formatador
	Imposto() CalculadoraImposto
	// Factory Method dentro da fabrica: o meio de pagamento por nome.
	Pagamento(meio string) (MeioPagamento, error)
}

// ---- Familia Brasil ----

type docBR struct{}

var reCPF = regexp.MustCompile(`^\d{3}\.\d{3}\.\d{3}-\d{2}$`)

func (docBR) Validar(d string) error {
	if !reCPF.MatchString(d) {
		return fmt.Errorf("CPF invalido: %q (esperado 000.000.000-00)", d)
	}
	return nil
}

type fmtBR struct{}

func (fmtBR) Moeda(c int64) string {
	reais, cent := c/100, c%100
	s := fmt.Sprintf("%d", reais)
	var partes []string
	for len(s) > 3 {
		partes = append([]string{s[len(s)-3:]}, partes...)
		s = s[:len(s)-3]
	}
	partes = append([]string{s}, partes...)
	return fmt.Sprintf("R$ %s,%02d", strings.Join(partes, "."), cent)
}

type impBR struct{}

func (impBR) Imposto(c int64) int64 { return c * 17 / 100 } // aliquota ilustrativa

type pixBR struct{}

func (pixBR) Autorizar(c int64) string { return "PIX instantaneo de " + fmtBR{}.Moeda(c) }

type boletoBR struct{}

func (boletoBR) Autorizar(c int64) string {
	return "boleto de " + fmtBR{}.Moeda(c) + " (compensa em 1-3 dias uteis)"
}

type RegiaoBR struct{}

func (RegiaoBR) Nome() string                  { return "BR" }
func (RegiaoBR) Documento() ValidadorDocumento { return docBR{} }
func (RegiaoBR) Formato() Formatador           { return fmtBR{} }
func (RegiaoBR) Imposto() CalculadoraImposto   { return impBR{} }
func (RegiaoBR) Pagamento(m string) (MeioPagamento, error) {
	switch m {
	case "pix":
		return pixBR{}, nil
	case "boleto":
		return boletoBR{}, nil
	}
	return nil, fmt.Errorf("meio %q indisponivel no BR", m)
}

// ---- Familia EUA ----

type docUS struct{}

var reSSN = regexp.MustCompile(`^\d{3}-\d{2}-\d{4}$`)

func (docUS) Validar(d string) error {
	if !reSSN.MatchString(d) {
		return fmt.Errorf("SSN invalido: %q (esperado 000-00-0000)", d)
	}
	return nil
}

type fmtUS struct{}

func (fmtUS) Moeda(c int64) string {
	d := c / 100
	s := fmt.Sprintf("%d", d)
	var partes []string
	for len(s) > 3 {
		partes = append([]string{s[len(s)-3:]}, partes...)
		s = s[:len(s)-3]
	}
	partes = append([]string{s}, partes...)
	return fmt.Sprintf("$%s.%02d", strings.Join(partes, ","), c%100)
}

type impUS struct{}

func (impUS) Imposto(c int64) int64 { return c * 8 / 100 }

type cartaoUS struct{}

func (cartaoUS) Autorizar(c int64) string { return "card authorized for " + fmtUS{}.Moeda(c) }

type RegiaoUS struct{}

func (RegiaoUS) Nome() string                  { return "US" }
func (RegiaoUS) Documento() ValidadorDocumento { return docUS{} }
func (RegiaoUS) Formato() Formatador           { return fmtUS{} }
func (RegiaoUS) Imposto() CalculadoraImposto   { return impUS{} }
func (RegiaoUS) Pagamento(m string) (MeioPagamento, error) {
	if m == "cartao" {
		return cartaoUS{}, nil
	}
	return nil, fmt.Errorf("meio %q indisponivel nos US", m)
}

// ---- Escolha da familia a partir de um codigo (configuracao, cabecalho, perfil) ----

func RegiaoPorCodigo(codigo string) (Regiao, error) {
	switch strings.ToUpper(codigo) {
	case "BR":
		return RegiaoBR{}, nil
	case "US":
		return RegiaoUS{}, nil
	}
	return nil, fmt.Errorf("regiao desconhecida: %q", codigo)
}

// ---- O cliente: o fechamento NAO sabe que regiao e ----

type Resultado struct {
	Total, Imposto string
	Pagamento      string
}

func Fechar(r Regiao, documento, meio string, subtotal int64) (Resultado, error) {
	if err := r.Documento().Validar(documento); err != nil {
		return Resultado{}, err
	}
	pg, err := r.Pagamento(meio)
	if err != nil {
		return Resultado{}, err
	}
	imp := r.Imposto().Imposto(subtotal)
	total := subtotal + imp
	f := r.Formato()
	return Resultado{Total: f.Moeda(total), Imposto: f.Moeda(imp), Pagamento: pg.Autorizar(total)}, nil
}
