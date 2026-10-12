package pedido

import "strconv"

// refDe monta a referência de cobrança de um pedido.
func refDe(p Pedido) string { return "pedido-" + strconv.Itoa(p.ID) }
