package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valkey-io/valkey-go"
	"golang.org/x/sync/singleflight"
)

func medir(nome string, f func()) {
	ini := time.Now()
	f()
	fmt.Printf("%-44s %s\n", nome, time.Since(ini).Round(100*time.Microsecond))
}

// buscarNoBanco simula uma consulta lenta (20 ms) e conta quantas vezes foi chamada.
var consultas atomic.Int64

func buscarNoBanco(id int) string {
	consultas.Add(1)
	time.Sleep(20 * time.Millisecond)
	return "produto-" + strconv.Itoa(id)
}

func main() {
	ctx := context.Background()
	c, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{"127.0.0.1:56390"}})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	c.Do(ctx, c.B().Flushall().Build())

	// ---- 1. uma ida e volta por comando x pipeline ----
	const N = 1000
	medir("1000 SET, um por vez (1000 idas e voltas)", func() {
		for i := range N {
			c.Do(ctx, c.B().Set().Key("k:"+strconv.Itoa(i)).Value("v").Build())
		}
	})
	medir("1000 SET em pipeline (DoMulti, 1 envio)", func() {
		cmds := make(valkey.Commands, 0, N)
		for i := range N {
			cmds = append(cmds, c.B().Set().Key("p:"+strconv.Itoa(i)).Value("v").Build())
		}
		for _, r := range c.DoMulti(ctx, cmds...) {
			if r.Error() != nil {
				log.Fatal(r.Error())
			}
		}
	})
	medir("1000 SET concorrentes (50 goroutines, auto-pipeline)", func() {
		var wg sync.WaitGroup
		for g := range 50 {
			wg.Go(func() {
				for i := range N / 50 {
					c.Do(ctx, c.B().Set().Key("c:"+strconv.Itoa(g*1000+i)).Value("v").Build())
				}
			})
		}
		wg.Wait()
	})

	// ---- 2. cache-aside: banco lento por tras de um cache ----
	buscar := func(id int) string {
		chave := "cache:produto:" + strconv.Itoa(id)
		if v, err := c.Do(ctx, c.B().Get().Key(chave).Build()).ToString(); err == nil {
			return v // acerto
		}
		v := buscarNoBanco(id) // erro de cache: vai ao banco...
		c.Do(ctx, c.B().Set().Key(chave).Value(v).Ex(time.Minute).Build())
		return v // ...e guarda
	}
	fmt.Println()
	medir("cache-aside: 1a leitura (vai ao banco)", func() { buscar(7) })
	medir("cache-aside: 2a leitura (acerto)", func() { buscar(7) })
	fmt.Println("  consultas ao banco ate aqui:", consultas.Load())

	// ---- 3. estouro de cache (stampede): 100 pedidos simultaneos de uma chave fria ----
	consultas.Store(0)
	c.Do(ctx, c.B().Del().Key("cache:produto:9").Build())
	var wg sync.WaitGroup
	medir("sem protecao: 100 leituras simultaneas, chave fria", func() {
		for range 100 {
			wg.Go(func() { buscar(9) })
		}
		wg.Wait()
	})
	fmt.Println("  consultas ao banco:", consultas.Load(), "<- todas foram ao banco")

	consultas.Store(0)
	c.Do(ctx, c.B().Del().Key("cache:produto:9").Build())
	var sf singleflight.Group
	medir("singleflight: 100 leituras simultaneas, chave fria", func() {
		for range 100 {
			wg.Go(func() {
				sf.Do("9", func() (any, error) { return buscar(9), nil })
			})
		}
		wg.Wait()
	})
	fmt.Println("  consultas ao banco:", consultas.Load(), "<- uma so, as outras esperaram")
}
