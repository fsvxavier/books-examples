// Package bce mostra a eliminação de verificações de limites.
package bce

// SomaIndice verifica o limite a cada acesso.
func SomaIndice(s []int, idx []int) int {
	t := 0
	for _, i := range idx {
		t += s[i]
	}
	return t
}

// SomaRange percorre com range: o compilador prova que está dentro.
func SomaRange(s []int) int {
	t := 0
	for _, v := range s {
		t += v
	}
	return t
}

// Quatro lê quatro posições: uma dica elimina três verificações.
func Quatro(s []byte) uint32 {
	return uint32(s[0]) | uint32(s[1])<<8 | uint32(s[2])<<16 | uint32(s[3])<<24
}

// QuatroComDica faz uma verificação só, no início.
func QuatroComDica(s []byte) uint32 {
	_ = s[3]
	return uint32(s[0]) | uint32(s[1])<<8 | uint32(s[2])<<16 | uint32(s[3])<<24
}

// Janela usa o laço clássico com len(s): também provado.
func Janela(s []int) int {
	t := 0
	for i := 0; i < len(s)-1; i++ {
		t += s[i] * s[i+1]
	}
	return t
}
