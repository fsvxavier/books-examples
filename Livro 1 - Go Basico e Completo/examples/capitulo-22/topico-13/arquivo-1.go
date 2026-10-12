package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// Chaveiro guarda as chaves-mestras (KEK) por versao. Em producao, ficam em um
// KMS ou HSM; aqui, em memoria, para o exemplo.
type Chaveiro struct {
	atual  int
	keks   map[int][]byte
	indice []byte // chave do indice cego
}

func NovoChaveiro() *Chaveiro {
	k := &Chaveiro{atual: 1, keks: map[int][]byte{1: aleatorio(32)}, indice: aleatorio(32)}
	return k
}

func (c *Chaveiro) Rotacionar() { c.atual++; c.keks[c.atual] = aleatorio(32) }

func aleatorio(n int) []byte { b := make([]byte, n); rand.Read(b); return b }

func gcm(chave []byte) cipher.AEAD {
	b, _ := aes.NewCipher(chave)
	g, _ := cipher.NewGCMWithRandomNonce(b)
	return g
}

// Registro e o que vai para o banco: nada aqui revela o CPF.
type Registro struct {
	ID         int
	VersaoKEK  int
	DEKCifrada []byte // chave de dados, cifrada pela KEK
	CPFCifrado []byte
	CPFIndice  string // HMAC do CPF: permite buscar sem decifrar
}

var ErrDecifrar = errors.New("nao foi possivel decifrar")

func (c *Chaveiro) Proteger(id int, cpf string) Registro {
	dek := aleatorio(32) // uma chave nova por registro
	ad := []byte(fmt.Sprintf("registro:%d", id))
	m := hmac.New(sha256.New, c.indice)
	m.Write([]byte(cpf))
	return Registro{
		ID:         id,
		VersaoKEK:  c.atual,
		DEKCifrada: gcm(c.keks[c.atual]).Seal(nil, nil, dek, ad),
		CPFCifrado: gcm(dek).Seal(nil, nil, []byte(cpf), ad),
		CPFIndice:  hex.EncodeToString(m.Sum(nil)),
	}
}

func (c *Chaveiro) Revelar(r Registro) (string, error) {
	ad := []byte(fmt.Sprintf("registro:%d", r.ID))
	dek, err := gcm(c.keks[r.VersaoKEK]).Open(nil, nil, r.DEKCifrada, ad)
	if err != nil {
		return "", ErrDecifrar
	}
	cpf, err := gcm(dek).Open(nil, nil, r.CPFCifrado, ad)
	if err != nil {
		return "", ErrDecifrar
	}
	return string(cpf), nil
}

// Reembrulhar troca a KEK SEM tocar nos dados: so a DEK e recifrada.
func (c *Chaveiro) Reembrulhar(r Registro) (Registro, error) {
	ad := []byte(fmt.Sprintf("registro:%d", r.ID))
	dek, err := gcm(c.keks[r.VersaoKEK]).Open(nil, nil, r.DEKCifrada, ad)
	if err != nil {
		return r, ErrDecifrar
	}
	r.DEKCifrada = gcm(c.keks[c.atual]).Seal(nil, nil, dek, ad)
	r.VersaoKEK = c.atual
	return r, nil
}

func main() {
	c := NovoChaveiro()
	r1 := c.Proteger(1, "123.456.789-09")
	r2 := c.Proteger(2, "123.456.789-09")

	fmt.Printf("registro 1: kek=v%d cpf cifrado=%x...\n", r1.VersaoKEK, r1.CPFCifrado[:10])
	fmt.Printf("registro 2: kek=v%d cpf cifrado=%x...\n", r2.VersaoKEK, r2.CPFCifrado[:10])
	fmt.Println("mesmo CPF, cifrados diferentes:", string(r1.CPFCifrado) != string(r2.CPFCifrado))
	fmt.Println("mesmo CPF, mesmo indice cego:  ", r1.CPFIndice == r2.CPFIndice, "(permite WHERE cpf_indice = ...)")

	cpf, err := c.Revelar(r1)
	fmt.Printf("revelar 1: %q err=%v\n", cpf, err)

	// atacante copia o CPF cifrado do registro 1 para o registro 2: o AD nao bate
	trocado := r2
	trocado.DEKCifrada, trocado.CPFCifrado = r1.DEKCifrada, r1.CPFCifrado
	_, err = c.Revelar(trocado)
	fmt.Println("dados movidos de linha:", err)

	// rotacao da chave-mestra
	c.Rotacionar()
	novo, err := c.Reembrulhar(r1)
	fmt.Printf("\nreembrulhado: kek=v%d (dados cifrados identicos: %v) err=%v\n",
		novo.VersaoKEK, string(novo.CPFCifrado) == string(r1.CPFCifrado), err)
	cpf, err = c.Revelar(novo)
	fmt.Printf("revelar apos rotacao: %q err=%v\n", cpf, err)
	cpf, err = c.Revelar(r2) // ainda na KEK v1, que continua guardada
	fmt.Printf("registro 2 ainda na v1: %q err=%v\n", cpf, err)
}
