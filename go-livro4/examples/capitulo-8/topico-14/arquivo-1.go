// Command e09 sobe um serviço go-micro v6 com um método Total e o chama por um cliente go-micro.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	micro "go-micro.dev/v6"
	"go-micro.dev/v6/registry"
)

type Pedido struct{ Codigo string }
type Resposta struct{ Acessos int }

type Estatisticas struct{}

func (Estatisticas) Total(ctx context.Context, req *Pedido, rsp *Resposta) error {
	rsp.Acessos = len(req.Codigo)
	return nil
}

func main() {
	reg := registry.NewMemoryRegistry()
	svc := micro.New("estatisticas",
		micro.Address("127.0.0.1:0"),
		micro.Registry(reg),
	)
	if err := svc.Handle(Estatisticas{}); err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := svc.Run(); err != nil {
			log.Println("serviço:", err)
		}
	}()
	time.Sleep(time.Second)

	var rsp Resposta
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req := svc.Client().NewRequest("estatisticas", "Estatisticas.Total", &Pedido{Codigo: "abc234"})
	if err := svc.Client().Call(ctx, req, &rsp); err != nil {
		log.Fatal(err)
	}
	fmt.Println("acessos:", rsp.Acessos)
}
