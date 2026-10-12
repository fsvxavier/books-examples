// pagamento/gateway.go
package pagamento

import "fmt"

// Gateway não conhece o pacote pedido: recebe só dados simples.
type Gateway struct{}

func (g *Gateway) Cobrar(ref string, valor int64) error {
	fmt.Printf("cobrando %d centavos (%s)\n", valor, ref)
	return nil
}
