package main

import "fmt"

func main() {
	// Uma arvore de regras para um e-mail: (nao vazio E tamanho E (contem @ E nao contem espacos))
	email := Todas(
		NaoVazio("email"),
		Tamanho("email", 5, 40),
		Contem("email", "@"),
		Nao(Contem("email", " "), "email: nao pode ter espacos"),
	)

	// Uma regra de documento: CPF (11) OU CNPJ (14), com tamanho como unica folha.
	documento := Alguma(Tamanho("cpf", 11, 11), Tamanho("cnpj", 14, 14))

	for _, e := range []string{"ana@exemplo.com", "ana exemplo", "", "a@b"} {
		fmt.Printf("%q:\n", e)
		if err := email.Validar(e); err != nil {
			fmt.Println(indentar(err.Error()))
		} else {
			fmt.Println("  ok")
		}
	}

	for _, d := range []string{"12345678901", "12345678000199", "123"} {
		err := documento.Validar(d)
		fmt.Printf("documento %q: %v\n", d, err != nil)
	}
	fmt.Println("detalhe do erro do documento 123:")
	fmt.Println(indentar(documento.Validar("123").Error()))
}

func indentar(s string) string {
	out := "  "
	for _, r := range s {
		out += string(r)
		if r == '\n' {
			out += "  "
		}
	}
	return out
}
