package main

import "fmt"

func main() {
	caneta := Produto{"Caneta", 1000}
	caderno := Produto{"Caderno", 3000}
	mochila := Produto{"Mochila", 20000}

	kit := NovoKit("kit escolar", caneta, caderno)
	pedido := NovoKit("pedido", kit, mochila)
	fmt.Println(pedido.Descricao(), "=", pedido.Preco())

	// Decorator em UM no da arvore (so o kit ganha desconto).
	kitComDesconto := DescontoFixo(kit, 500)
	pedido2 := NovoKit("pedido", kitComDesconto, mochila)
	fmt.Println(pedido2.Descricao(), "=", pedido2.Preco())

	// Decorator na RAIZ: imposto sobre tudo. A arvore continua sendo um Item.
	fmt.Println("com imposto de 10% na raiz:", Imposto(pedido2, 1000).Preco())

	// A ORDEM dos decoradores muda o resultado (desconto fixo e imposto nao comutam).
	a := Imposto(DescontoFixo(pedido, 3000), 1000) // desconta, depois aplica imposto
	b := DescontoFixo(Imposto(pedido, 1000), 3000) // aplica imposto, depois desconta
	fmt.Println("imposto sobre (preco - desconto):", a.Preco(), "|", a.Descricao())
	fmt.Println("(preco com imposto) - desconto:   ", b.Preco(), "|", b.Descricao())

	// Desconto nao leva o preco abaixo de zero.
	fmt.Println("desconto maior que o preco:", DescontoFixo(caneta, 99999).Preco())

	// Cache decorando uma SUBARVORE: ganho e risco.
	cache := &ComCache{Item: kit}
	grande := NovoKit("grande", cache, cache, cache)
	fmt.Println("3 usos do mesmo kit em cache:", grande.Preco(), "| calculos:", cache.calculos.Load())
	kit.Adicionar(Produto{"Lapis", 200}) // muda a arvore por baixo do cache
	fmt.Println("depois de adicionar um lapis ao kit (cache velho):", grande.Preco(), "| preco real de 3 kits:", 3*kit.Preco())
	cache.Invalidar()
	fmt.Println("depois de Invalidar:", grande.Preco(), "| calculos:", cache.calculos.Load())
}
