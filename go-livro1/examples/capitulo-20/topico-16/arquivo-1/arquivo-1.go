package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/valkey-io/valkey-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	c, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{"127.0.0.1:56390"}})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	if err := c.Do(ctx, c.B().Flushall().Build()).Error(); err != nil {
		log.Fatal(err)
	}

	// Create / Read / Update / Delete de uma chave
	c.Do(ctx, c.B().Set().Key("produto:1:nome").Value("Teclado").Build())
	nome, _ := c.Do(ctx, c.B().Get().Key("produto:1:nome").Build()).ToString()
	fmt.Println("GET:", nome)

	c.Do(ctx, c.B().Set().Key("produto:1:nome").Value("Teclado Mecanico").Build()) // update = SET de novo
	nome, _ = c.Do(ctx, c.B().Get().Key("produto:1:nome").Build()).ToString()
	fmt.Println("apos update:", nome)

	n, _ := c.Do(ctx, c.B().Del().Key("produto:1:nome").Build()).AsInt64()
	fmt.Println("DEL removeu:", n)
	_, err = c.Do(ctx, c.B().Get().Key("produto:1:nome").Build()).ToString()
	fmt.Println("GET apos DEL: nil =", valkey.IsValkeyNil(err), "|", err)

	// TTL: a chave some sozinha
	c.Do(ctx, c.B().Set().Key("sessao:abc").Value("ana").Ex(2*time.Second).Build())
	ttl, _ := c.Do(ctx, c.B().Ttl().Key("sessao:abc").Build()).AsInt64()
	fmt.Println("TTL da sessao (s):", ttl)
	time.Sleep(2100 * time.Millisecond)
	_, err = c.Do(ctx, c.B().Get().Key("sessao:abc").Build()).ToString()
	fmt.Println("apos 2,1s: expirou =", valkey.IsValkeyNil(err))

	// contador atomico
	for range 5 {
		c.Do(ctx, c.B().Incr().Key("visitas").Build())
	}
	v, _ := c.Do(ctx, c.B().Get().Key("visitas").Build()).AsInt64()
	fmt.Println("INCR x5:", v)

	// hash: um registro com campos
	c.Do(ctx, c.B().Hset().Key("produto:2").FieldValue().FieldValue("nome", "Mouse").FieldValue("preco", "8990").Build())
	h, _ := c.Do(ctx, c.B().Hgetall().Key("produto:2").Build()).AsStrMap()
	fmt.Println("HGETALL:", h)

	// sorted set: ranking
	for nome, pontos := range map[string]float64{"ana": 120, "bia": 95, "caio": 140} {
		c.Do(ctx, c.B().Zadd().Key("ranking").ScoreMember().ScoreMember(pontos, nome).Build())
	}
	top, _ := c.Do(ctx, c.B().Zrevrange().Key("ranking").Start(0).Stop(1).Build()).AsStrSlice()
	fmt.Println("top 2 do ranking:", top)

	// SET NX: trava simples (so um vence)
	a, _ := c.Do(ctx, c.B().Set().Key("trava").Value("a").Nx().Ex(5*time.Second).Build()).ToString()
	_, errB := c.Do(ctx, c.B().Set().Key("trava").Value("b").Nx().Ex(5*time.Second).Build()).ToString()
	fmt.Println("trava: A pegou =", a == "OK", "| B pegou =", !valkey.IsValkeyNil(errB))
}
