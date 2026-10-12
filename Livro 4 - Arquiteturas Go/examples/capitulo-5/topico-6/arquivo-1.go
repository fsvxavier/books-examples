// Comando e10: mede o que acontece quando um contador de acessos mora dentro
// de um agregado protegido por bloqueio otimista (versão).
package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var errConflito = errors.New("conflito de versão")

// linha é o que o "banco" guarda: o agregado inteiro e a sua versão.
type linha struct {
	versao  int
	acessos int
}

// banco imita um repositório com atualização condicional:
// UPDATE ... SET acessos=$1, versao=versao+1 WHERE versao=$2.
type banco struct {
	mu    sync.Mutex
	dados linha
}

func (b *banco) ler() linha {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.dados
}

func (b *banco) gravar(nova linha, versaoLida int) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.dados.versao != versaoLida {
		return errConflito
	}
	nova.versao = versaoLida + 1
	b.dados = nova
	return nil
}

// acessar é o caso de uso ingênuo: carrega o agregado, aplica a regra de
// negócio, grava de volta, e tenta de novo se alguém gravou antes.
// ida e volta simula o tempo de uma consulta ao banco pela rede.
func acessar(b *banco, tentativas int, idaEVolta time.Duration, reenvios *atomic.Int64) error {
	for i := 0; i < tentativas; i++ {
		l := b.ler()
		time.Sleep(idaEVolta) // a rede
		l.acessos++           // a "regra" do agregado
		err := b.gravar(l, l.versao)
		if err == nil {
			return nil
		}
		reenvios.Add(1)
	}
	return errConflito
}

func rodada(concorrentes, tentativas int, idaEVolta time.Duration) (ok, falhas int, reenvios int64, total int) {
	b := &banco{}
	var r atomic.Int64
	var wg sync.WaitGroup
	var oks, fs atomic.Int64
	for range concorrentes {
		wg.Go(func() {
			if acessar(b, tentativas, idaEVolta, &r) == nil {
				oks.Add(1)
			} else {
				fs.Add(1)
			}
		})
	}
	wg.Wait()
	return int(oks.Load()), int(fs.Load()), r.Load(), b.ler().acessos
}

// incrementoAtomico é a alternativa: a regra "somar um" vai para o banco, em
// uma instrução só (UPDATE ... SET acessos = acessos + 1), sem ler antes.
func incrementoAtomico(concorrentes int, idaEVolta time.Duration) (ok, total int) {
	var contador atomic.Int64
	var wg sync.WaitGroup
	for range concorrentes {
		wg.Go(func() {
			time.Sleep(idaEVolta) // a mesma rede
			contador.Add(1)
		})
	}
	wg.Wait()
	return concorrentes, int(contador.Load())
}

func main() {
	const idaEVolta = 200 * time.Microsecond
	fmt.Printf("%-13s %-10s %-9s %-8s %-10s %s\n", "concorrentes", "tentativas", "sucessos", "falhas", "reenvios", "contador final")
	for _, c := range []int{10, 50, 100, 500} {
		for _, t := range []int{5, 50} {
			ok, f, r, total := rodada(c, t, idaEVolta)
			fmt.Printf("%-13d %-10d %-9d %-8d %-10d %d\n", c, t, ok, f, r, total)
		}
	}
	ok, total := incrementoAtomico(500, idaEVolta)
	fmt.Printf("\ncontador atômico: 500 concorrentes, %d sucessos, 0 falhas, 0 reenvios, contador final %d\n", ok, total)
}
