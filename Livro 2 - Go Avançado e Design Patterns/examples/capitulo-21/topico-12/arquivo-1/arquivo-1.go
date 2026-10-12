package inl

import "sync"

var mu sync.Mutex

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func SomaLaco(s []int) int {
	t := 0
	for _, v := range s {
		t += v
	}
	return t
}

func ComDefer() int {
	mu.Lock()
	defer mu.Unlock()
	return 1
}

func Grande(a, b, c int) int {
	x := a*b + c
	x += a*c - b
	x ^= b * 31
	x += (a + b) * (b + c) * (a + c)
	if x > 100 {
		x -= a
	} else {
		x += b
	}
	for i := 0; i < c; i++ {
		x += i*a - b
		if x%7 == 0 {
			x++
		}
	}
	switch {
	case x > 1000:
		x /= 3
	case x < -1000:
		x *= 2
	}
	return x*7 + a*3 - c*11
}

func ChamaAbs(v int) int { return Abs(v) + 1 }

// Vizinhos soma s[i] com o proximo; o compilador nao prova que i+1 cabe.
func Vizinhos(s []int) int {
	t := 0
	for i := 0; i < len(s)-1; i++ {
		t += s[i] + s[i+1]
	}
	return t
}

// VizinhosSemChecagem avisa o compilador do tamanho antes do laco.
func VizinhosSemChecagem(s []int) int {
	t := 0
	for i := 1; i < len(s); i++ {
		t += s[i-1] + s[i]
	}
	return t
}

// Pegar nao tem como provar que i esta dentro do slice: a checagem fica.
func Pegar(s []int, i int) int { return s[i] }

// Pular2 le tres posicoes seguidas; so a primeira checagem e necessaria.
func Pular2(s []int) int {
	t := 0
	for i := 0; i+2 < len(s); i++ {
		t += s[i] + s[i+1] + s[i+2]
	}
	return t
}

// PularUsandoI le s[i+2] com a condicao errada para o compilador.
func PularErrado(s []int) int {
	t := 0
	for i := 0; i < len(s)-1; i++ {
		t += s[i] + s[i+2]
	}
	return t
}
