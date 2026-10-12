package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/valkey-io/valkey-go"
)

func main() {
	ctx := context.Background()
	opt := valkey.ClientOption{InitAddress: []string{"127.0.0.1:56390"}}
	a, err := valkey.NewClient(opt)
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()
	b, _ := valkey.NewClient(opt) // "outro servico"
	defer b.Close()

	a.Do(ctx, a.B().Set().Key("config:tema").Value("claro").Build())

	// ---- Cache no cliente: DoCache guarda a resposta na memoria do processo ----
	leitura := func() (string, time.Duration) {
		ini := time.Now()
		v, _ := a.DoCache(ctx, a.B().Get().Key("config:tema").Cache(), time.Minute).ToString()
		return v, time.Since(ini)
	}
	v, d := leitura()
	fmt.Printf("1a leitura (vai ao servidor):   %q em %s\n", v, d.Round(time.Microsecond))
	v, d = leitura()
	fmt.Printf("2a leitura (cache no cliente):  %q em %s\n", v, d.Round(time.Microsecond))

	// outro cliente altera: o servidor AVISA o primeiro, que descarta o valor velho
	b.Do(ctx, b.B().Set().Key("config:tema").Value("escuro").Build())
	time.Sleep(50 * time.Millisecond)
	v, d = leitura()
	fmt.Printf("apos outro cliente alterar:     %q em %s (invalidado e relido)\n", v, d.Round(time.Microsecond))

	// ---- Pub/Sub: avisos em tempo real ----
	recebidos := make(chan string, 3)
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go a.Receive(subCtx, a.B().Subscribe().Channel("avisos").Build(), func(m valkey.PubSubMessage) {
		recebidos <- m.Message
	})
	time.Sleep(100 * time.Millisecond) // deixa a assinatura se firmar
	for _, msg := range []string{"pedido 1", "pedido 2"} {
		b.Do(ctx, b.B().Publish().Channel("avisos").Message(msg).Build())
	}
	for range 2 {
		select {
		case m := <-recebidos:
			fmt.Println("Pub/Sub recebeu:", m)
		case <-time.After(2 * time.Second):
			fmt.Println("Pub/Sub: nada recebido")
		}
	}

	// ---- O protocolo e o mesmo: um cliente go-redis tambem fala com o Valkey ----
	r := redis.NewClient(&redis.Options{Addr: "127.0.0.1:56390"})
	defer r.Close()
	tema, err := r.Get(ctx, "config:tema").Result()
	fmt.Println("go-redis lendo do Valkey:", tema, err)
}
