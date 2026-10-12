package e06

import "net/mail"

// Endereco é uma struct aninhada.
type Endereco struct {
	Rua string `validate:"obrigatorio"`
	CEP string `validate:"obrigatorio,min=8,max=9"`
}

// Contato é validado elemento a elemento (dive).
type Contato struct {
	Tipo string `validate:"obrigatorio"`
	Fone string `validate:"min=8"`
}

// Cadastro reúne as regras do validador.
type Cadastro struct {
	Nome     string    `validate:"obrigatorio,min=3,max=20"`
	Idade    int       `validate:"min=18,max=130"`
	Email    string    `validate:"obrigatorio,email"`
	Senha    string    `validate:"-"`
	Tags     []string  `validate:"max=3"`
	Endereco Endereco  // validado pelos próprios campos
	Cobranca *Endereco // validado só se não for nulo
	Contatos []Contato `validate:"dive"`
	interno  string    // não exportado: ignorado
}

// ValidarAMao é a mesma validação escrita sem reflexão, só para comparar custo e comportamento.
func ValidarAMao(c *Cadastro) Falhas {
	var fs Falhas
	add := func(campo, regra, msg string) { fs = append(fs, Falha{campo, regra, msg}) }
	n := len([]rune(c.Nome))
	if c.Nome == "" {
		add("Nome", "obrigatorio", "é obrigatório")
	}
	if float64(n) < 3 {
		add("Nome", "min", "")
	}
	if n > 20 {
		add("Nome", "max", "")
	}
	if c.Idade < 18 {
		add("Idade", "min", "")
	}
	if c.Idade > 130 {
		add("Idade", "max", "")
	}
	if c.Email == "" {
		add("Email", "obrigatorio", "é obrigatório")
	} else if a, err := mail.ParseAddress(c.Email); err != nil || a.Address != c.Email {
		add("Email", "email", "não é um e-mail válido")
	}
	if len(c.Tags) > 3 {
		add("Tags", "max", "")
	}
	valida := func(prefixo string, e *Endereco) {
		if e.Rua == "" {
			add(prefixo+".Rua", "obrigatorio", "")
		}
		if e.CEP == "" {
			add(prefixo+".CEP", "obrigatorio", "")
		}
		if l := len([]rune(e.CEP)); l < 8 {
			add(prefixo+".CEP", "min", "")
		} else if l > 9 {
			add(prefixo+".CEP", "max", "")
		}
	}
	valida("Endereco", &c.Endereco)
	if c.Cobranca != nil {
		valida("Cobranca", c.Cobranca)
	}
	for i, ct := range c.Contatos {
		if ct.Tipo == "" {
			add("Contatos["+itoa(i)+"].Tipo", "obrigatorio", "")
		}
		if len([]rune(ct.Fone)) < 8 {
			add("Contatos["+itoa(i)+"].Fone", "min", "")
		}
	}
	return fs
}

func itoa(i int) string { return string(rune('0' + i)) }
