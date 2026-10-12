package main

import "fmt"

// Pilha: último a entrar, primeiro a sair. O valor zero já é uma pilha vazia.
type Pilha struct{ itens []int }

func (p *Pilha) Empilhar(v int) { p.itens = append(p.itens, v) }

func (p *Pilha) Desempilhar() (int, bool) {
	if len(p.itens) == 0 {
		return 0, false
	}
	topo := len(p.itens) - 1
	v := p.itens[topo]
	p.itens = p.itens[:topo]
	return v, true
}

func (p *Pilha) Tamanho() int { return len(p.itens) }

// FilaSimples remove do início com fatiamento: os elementos removidos continuam
// no array até que um append force a realocação.
type FilaSimples struct{ itens []int }

func (f *FilaSimples) Enfileirar(v int) { f.itens = append(f.itens, v) }
func (f *FilaSimples) Desenfileirar() (int, bool) {
	if len(f.itens) == 0 {
		return 0, false
	}
	v := f.itens[0]
	f.itens = f.itens[1:]
	return v, true
}

// Fila usa um índice de início e compacta o array quando metade dele já foi consumida.
type Fila struct {
	itens  []int
	inicio int
}

func (f *Fila) Enfileirar(v int) { f.itens = append(f.itens, v) }

func (f *Fila) Desenfileirar() (int, bool) {
	if f.inicio >= len(f.itens) {
		return 0, false
	}
	v := f.itens[f.inicio]
	f.inicio++
	if f.inicio > len(f.itens)/2 {
		f.itens = append([]int(nil), f.itens[f.inicio:]...)
		f.inicio = 0
	}
	return v, true
}

func (f *Fila) Tamanho() int { return len(f.itens) - f.inicio }

// balanceada usa a pilha para verificar parênteses, colchetes e chaves.
func balanceada(s string) bool {
	pares := map[rune]rune{')': '(', ']': '[', '}': '{'}
	var abertos []rune
	for _, r := range s {
		switch r {
		case '(', '[', '{':
			abertos = append(abertos, r)
		case ')', ']', '}':
			if len(abertos) == 0 || abertos[len(abertos)-1] != pares[r] {
				return false
			}
			abertos = abertos[:len(abertos)-1]
		}
	}
	return len(abertos) == 0
}

func main() {
	var p Pilha
	for _, v := range []int{1, 2, 3} {
		p.Empilhar(v)
	}
	for p.Tamanho() > 0 {
		v, _ := p.Desempilhar()
		fmt.Print(v, " ")
	}
	_, ok := p.Desempilhar()
	fmt.Println(ok)

	var f Fila
	for i := 1; i <= 6; i++ {
		f.Enfileirar(i)
	}
	for range 4 {
		v, _ := f.Desenfileirar()
		fmt.Print(v, " ")
	}
	fmt.Println(f.Tamanho(), len(f.itens), f.inicio)

	for _, s := range []string{"{[()]}", "([)]", "((", "", "a(b)c"} {
		fmt.Printf("%q: %t\n", s, balanceada(s))
	}
}
