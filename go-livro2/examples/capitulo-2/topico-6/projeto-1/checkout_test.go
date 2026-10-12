package checkout

import "testing"

func TestFechar(t *testing.T) {
	casos := []struct {
		nome, regiao, doc, meio string
		subtotal                int64
		total, imposto, pg      string
		erro                    bool
	}{
		{"BR pix", "BR", "123.456.789-09", "pix", 100000, "R$ 1.170,00", "R$ 170,00", "PIX instantaneo de R$ 1.170,00", false},
		{"BR boleto", "br", "123.456.789-09", "boleto", 5990, "R$ 70,08", "R$ 10,18", "boleto de R$ 70,08 (compensa em 1-3 dias uteis)", false},
		{"US cartao", "US", "123-45-6789", "cartao", 100000, "$1,080.00", "$80.00", "card authorized for $1,080.00", false},
		{"BR com SSN", "BR", "123-45-6789", "pix", 100, "", "", "", true},
		{"US com CPF", "US", "123.456.789-09", "cartao", 100, "", "", "", true},
		{"BR cartao nao existe", "BR", "123.456.789-09", "cartao", 100, "", "", "", true},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			r, err := RegiaoPorCodigo(c.regiao)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Fechar(r, c.doc, c.meio, c.subtotal)
			if c.erro {
				if err == nil {
					t.Fatalf("esperava erro, veio %+v", got)
				}
				return
			}
			if err != nil || got.Total != c.total || got.Imposto != c.imposto || got.Pagamento != c.pg {
				t.Fatalf("got %+v err=%v", got, err)
			}
		})
	}
	if _, err := RegiaoPorCodigo("XX"); err == nil {
		t.Fatal("regiao XX devia falhar")
	}
}
