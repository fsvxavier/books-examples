// Package pagamento implementa a cobrança. Não importa o pacote pedido.
package pagamento

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Gateway simula um provedor de pagamento externo.
type Gateway struct {
	Log   *slog.Logger
	Delay time.Duration
}

func (g *Gateway) Cobrar(ctx context.Context, ref string, valor int64) error {
	select {
	case <-time.After(g.Delay):
		g.Log.Info("cobrança aprovada", "ref", ref, "valor", valor)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("cobrança %s cancelada: %w", ref, context.Cause(ctx))
	}
}
