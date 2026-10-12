package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Pool simula um recurso caro por inquilino (tenant): uma conexao, um cache...
type Pool struct {
	Tenant string
	ID     int64
}

var abertos atomic.Int64

func abrir(tenant string) *Pool {
	time.Sleep(5 * time.Millisecond) // custo de abrir
	return &Pool{Tenant: tenant, ID: abertos.Add(1)}
}

// PorChave garante UMA instancia por chave, criada de forma preguicosa.
type PorChave[K comparable, V any] struct {
	novo func(K) V
	m    sync.Map // K -> func() V
}

func NovoPorChave[K comparable, V any](novo func(K) V) *PorChave[K, V] {
	return &PorChave[K, V]{novo: novo}
}

func (p *PorChave[K, V]) Obter(chave K) V {
	if f, ok := p.m.Load(chave); ok {
		return f.(func() V)()
	}
	// Guardamos a FUNCAO que cria, nao o valor: assim, mesmo que duas goroutines
	// cheguem juntas, so uma funcao vence o LoadOrStore e so ela executa.
	f, _ := p.m.LoadOrStore(chave, sync.OnceValue(func() V { return p.novo(chave) }))
	return f.(func() V)()
}

func main() {
	pools := NovoPorChave(abrir)
	tenants := []string{"acme", "globex", "initech"}

	var wg sync.WaitGroup
	vistos := make([]*Pool, 90)
	for i := range vistos {
		wg.Go(func() { vistos[i] = pools.Obter(tenants[i%3]) })
	}
	wg.Wait()

	porTenant := map[string]map[int64]bool{}
	for _, p := range vistos {
		if porTenant[p.Tenant] == nil {
			porTenant[p.Tenant] = map[int64]bool{}
		}
		porTenant[p.Tenant][p.ID] = true
	}
	fmt.Println("chamadas:", len(vistos))
	fmt.Println("pools abertos:", abertos.Load())
	for _, t := range tenants {
		fmt.Printf("tenant %-8s instancias distintas: %d\n", t, len(porTenant[t]))
	}
}
