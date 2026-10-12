package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

func main() {
	msg := []byte("transferir 100 para conta 42")

	// funcoes de hash: tamanho fixo, mesma entrada = mesma saida
	fmt.Printf("md5      %s  (quebrado: NAO use para seguranca)\n", sum(md5.Sum(msg)))
	s256 := sha256.Sum256(msg)
	fmt.Printf("sha256   %x\n", s256)
	s3 := sha3.Sum256(msg)
	fmt.Printf("sha3-256 %x\n", s3)

	// efeito avalanche: um bit muda tudo
	alt := append([]byte(nil), msg...)
	alt[len(alt)-1] ^= 1
	a2 := sha256.Sum256(alt)
	fmt.Printf("sha256 com 1 bit trocado %x\n", a2)
	diff := 0
	for i := range s256 {
		d := s256[i] ^ a2[i]
		for ; d != 0; d &= d - 1 {
			diff++
		}
	}
	fmt.Printf("bits diferentes: %d de 256\n", diff)

	// HMAC: hash com chave, autentica a mensagem
	chave := []byte("segredo-compartilhado-de-32-bytes!")
	mac := hmac.New(sha256.New, chave)
	mac.Write(msg)
	tag := mac.Sum(nil)
	fmt.Printf("hmac     %x\n", tag)

	verificar := func(msg, tag []byte) bool {
		m := hmac.New(sha256.New, chave)
		m.Write(msg)
		return hmac.Equal(m.Sum(nil), tag) // tempo constante
	}
	fmt.Println("msg original valida:", verificar(msg, tag))
	fmt.Println("msg adulterada valida:", verificar(alt, tag))

	// comparacao em tempo constante
	fmt.Println("subtle.ConstantTimeCompare:", subtle.ConstantTimeCompare(tag, tag))
}

func sum(b [16]byte) string { return hex.EncodeToString(b[:]) }
