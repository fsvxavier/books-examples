package main

import (
	"context"
	"fmt"
	"time"
)

// Instavel falha de forma temporaria nas primeiras N chamadas.
type Instavel struct {
	Canal
	falhas int
}

func (i *Instavel) Enviar(ctx context.Context, d, t string) error {
	if i.falhas > 0 {
		i.falhas--
		return fmt.Errorf("%w: gateway 503", ErrTemporario)
	}
	return i.Canal.Enviar(ctx, d, t)
}

func main() {
	ctx := context.Background()
	msg := "Pedido 1042 aprovado"

	fmt.Println("Simples sobre email, depois trocado para slack em tempo de execucao:")
	s := NovoSimples(Email{})
	s.Avisar(ctx, "ana@exemplo.com", msg)
	s.Trocar(Slack{})
	s.Avisar(ctx, "vendas", msg)

	fmt.Println("Simples sobre SMS com texto longo:")
	err := NovoSimples(SMS{}).Avisar(ctx, "+5511999990000", msg+" e ja esta a caminho do endereco cadastrado")
	fmt.Println("  erro:", err)

	fmt.Println("ComRetentativa sobre canal instavel (2 falhas):")
	r := ComRetentativa{Canal: &Instavel{Canal: Email{}, falhas: 2}, Tentativas: 4, Espera: 10 * time.Millisecond, Dormir: Dormir}
	fmt.Println("  resultado:", r.Avisar(ctx, "ana@exemplo.com", msg))

	fmt.Println("ComRetentativa que desiste (5 falhas, 3 tentativas):")
	r = ComRetentativa{Canal: &Instavel{Canal: Email{}, falhas: 5}, Tentativas: 3, Espera: time.Millisecond, Dormir: Dormir}
	fmt.Println("  resultado:", r.Avisar(ctx, "ana@exemplo.com", msg))

	fmt.Println("Urgente em tres canais (o SMS recusa o texto longo):")
	u := Urgente{
		Canais: []Canal{Email{}, SMS{}, Slack{}},
		Falhou: func(canal string, err error) { fmt.Printf("  aviso: %s falhou: %v\n", canal, err) },
	}
	fmt.Println("  resultado:", u.Avisar(ctx, "plantao", msg+" e tem prioridade"))
}
