package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	mrand "math/rand/v2"
	"slices"
	"time"
	"uuid"
)

// Token devolve n bytes aleatorios seguros, em base64 de URL (sem preenchimento).
func Token(n int) string {
	b := make([]byte, n)
	rand.Read(b) // desde o Go 1.24, nunca falha (ou o programa termina)
	return base64.RawURLEncoding.EncodeToString(b)
}

func main() {
	// crypto/rand: seguro
	fmt.Println("Token(32):  ", len(Token(32)), "caracteres,", "ex.", Token(8))
	fmt.Println("rand.Text():", rand.Text(), "(26 caracteres base32 = 128 bits)")
	n, _ := rand.Int(rand.Reader, big.NewInt(1_000_000))
	fmt.Println("rand.Int < 1.000.000:", n.Int64() < 1_000_000)

	// math/rand/v2 com a mesma semente: SEMPRE a mesma sequencia
	a := mrand.New(mrand.NewPCG(42, 1))
	b := mrand.New(mrand.NewPCG(42, 1))
	fmt.Println("math/rand/v2, semente 42:", a.Uint64(), a.Uint64(), "...")
	fmt.Println("math/rand/v2, semente 42:", b.Uint64(), b.Uint64(), "... (identica)")

	// token "aleatorio" semeado pelo relogio: o atacante adivinha a semente
	gerar := func(semente int64) string {
		r := mrand.New(mrand.NewPCG(uint64(semente), 0))
		return fmt.Sprintf("%016x", r.Uint64())
	}
	agora := time.Now().UnixNano()
	tokenVitima := gerar(agora)
	// atacante conhece o instante aproximado (janela de 1 milhao de ns = 1 ms)
	achou := false
	tentativas := 0
	for d := int64(-500_000); d <= 500_000; d++ {
		tentativas++
		if gerar(agora+d) == tokenVitima {
			achou = true
			break
		}
	}
	fmt.Printf("token fraco recuperado: %v em %d tentativas\n", achou, tentativas)

	// UUIDs (pacote uuid da biblioteca padrao, Go 1.27)
	u4 := uuid.New()
	fmt.Println("uuid.New():  ", u4)
	var v7s []string
	for range 5 {
		v7s = append(v7s, uuid.NewV7().String())
		time.Sleep(time.Millisecond)
	}
	fmt.Println("UUIDv7 em ordem de criacao esta ordenado:", slices.IsSorted(v7s))
	for _, s := range v7s[:2] {
		fmt.Println("  v7:", s)
	}
	u, err := uuid.Parse(v7s[0])
	fmt.Println("Parse:", u == uuid.MustParse(v7s[0]), err)
}
