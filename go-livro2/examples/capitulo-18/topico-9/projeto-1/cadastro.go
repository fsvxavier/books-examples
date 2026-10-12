// Package e07 usa um gerador de código no lugar da reflexão.
package e07

//go:generate go run ./gerar -arquivo modelo.go -saida validar_gen.go

// Cadastro é o modelo cujo validador é gerado.
type Cadastro struct {
	Nome  string `validate:"obrigatorio,min=3,max=20"`
	Idade int    `validate:"min=18,max=130"`
	Email string `validate:"obrigatorio,email"`
	Senha string `validate:"-"`
}

// Produto é outra struct do mesmo arquivo.
type Produto struct {
	SKU   string `validate:"obrigatorio,min=4"`
	Preco int    `validate:"min=1"`
}
