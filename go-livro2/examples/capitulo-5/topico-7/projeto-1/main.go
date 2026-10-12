package main

import (
	"context"
	"fmt"
	"time"

	"exemplos/livro2/capitulo-5/topico-7/projeto-1/frete"
)

func imprimir(ofertas []frete.Oferta, err error) {
	for _, o := range ofertas {
		fmt.Printf("  %-14s R$ %6.2f  %v\n", o.Transportadora, float64(o.Centavos)/100, o.Prazo)
	}
	fmt.Println("  erros:", err)
}

func main() {
	p := frete.Pedido{CEPOrigem: "01310100", CEPDestino: "20040020", PesoGramas: 2500}

	fmt.Println("todas respondem:")
	todas := []frete.Transportadora{frete.AdaptadorCorreios{}, frete.NovoRapidoLog(0), frete.NovaMoto(false)}
	imprimir(frete.Cotar(context.Background(), todas, p))

	fmt.Println("uma fora do ar, outra com CEP invalido:")
	falhas := []frete.Transportadora{frete.AdaptadorCorreios{}, frete.NovoRapidoLog(0), frete.NovaMoto(true)}
	ruim := p
	ruim.CEPDestino = "123"
	imprimir(frete.Cotar(context.Background(), falhas, ruim))

	fmt.Println("prazo de 50 ms e uma transportadora lenta (500 ms):")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	lentas := []frete.Transportadora{frete.AdaptadorCorreios{}, frete.NovoRapidoLog(500 * time.Millisecond)}
	inicio := time.Now()
	imprimir(frete.Cotar(ctx, lentas, p))
	fmt.Println("  demorou menos de 200 ms:", time.Since(inicio) < 200*time.Millisecond)
}
