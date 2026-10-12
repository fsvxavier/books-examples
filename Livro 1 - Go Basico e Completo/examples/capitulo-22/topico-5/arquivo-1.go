package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

func novoGCM(chave []byte) cipher.AEAD {
	bloco, err := aes.NewCipher(chave)
	if err != nil {
		panic(err)
	}
	g, err := cipher.NewGCM(bloco)
	if err != nil {
		panic(err)
	}
	return g
}

// Cifrar: nonce aleatorio de 12 bytes na frente do texto cifrado.
// ad (dados adicionais) sao autenticados, mas nao cifrados.
func Cifrar(g cipher.AEAD, texto, ad []byte) []byte {
	nonce := make([]byte, g.NonceSize())
	rand.Read(nonce)
	return g.Seal(nonce, nonce, texto, ad)
}

func Decifrar(g cipher.AEAD, cifrado, ad []byte) ([]byte, error) {
	if len(cifrado) < g.NonceSize() {
		return nil, fmt.Errorf("cifrado curto demais")
	}
	nonce, corpo := cifrado[:g.NonceSize()], cifrado[g.NonceSize():]
	return g.Open(nil, nonce, corpo, ad)
}

func main() {
	chave := make([]byte, 32) // AES-256
	rand.Read(chave)
	g := novoGCM(chave)

	texto := []byte("CPF 123.456.789-09")
	ad := []byte("usuario:42")

	c := Cifrar(g, texto, ad)
	fmt.Printf("texto: %d bytes, cifrado: %d bytes (12 nonce + %d texto + 16 etiqueta)\n", len(texto), len(c), len(texto))
	p, err := Decifrar(g, c, ad)
	fmt.Printf("decifrado: %q err=%v\n", p, err)

	// mesma entrada cifrada duas vezes: saidas diferentes (nonce aleatorio)
	fmt.Println("cifrar duas vezes da o mesmo resultado:", bytes.Equal(Cifrar(g, texto, ad), Cifrar(g, texto, ad)))

	// adulteracao e detectada
	adulterado := bytes.Clone(c)
	adulterado[20] ^= 0x01
	_, err = Decifrar(g, adulterado, ad)
	fmt.Println("1 bit alterado:", err)

	// dados adicionais errados: tambem falha (liga o cifrado a um contexto)
	_, err = Decifrar(g, c, []byte("usuario:99"))
	fmt.Println("ad errado:     ", err)

	// ERRO FATAL: reutilizar o nonce com a mesma chave
	nonce := make([]byte, g.NonceSize())
	a := []byte("ATAQUE AO AMANHECER")
	b := []byte("RECUAR IMEDIATAMENTE")
	ca := g.Seal(nil, nonce, a, nil)
	cb := g.Seal(nil, nonce, b, nil)
	n := min(len(a), len(b))
	xorC := make([]byte, n)
	xorP := make([]byte, n)
	for i := range n {
		xorC[i] = ca[i] ^ cb[i]
		xorP[i] = a[i] ^ b[i]
	}
	fmt.Println("nonce repetido: XOR dos cifrados == XOR dos textos:", bytes.Equal(xorC, xorP))

	// Go 1.24: GCM com nonce aleatorio embutido
	bloco, _ := aes.NewCipher(chave)
	gr, _ := cipher.NewGCMWithRandomNonce(bloco)
	cr := gr.Seal(nil, nil, texto, ad)
	pr, err := gr.Open(nil, nil, cr, ad)
	fmt.Printf("NewGCMWithRandomNonce: %d bytes, %q, err=%v\n", len(cr), pr, err)

	// ChaCha20-Poly1305 (rapido sem aceleracao de hardware)
	cc, _ := chacha20poly1305.New(chave)
	nc := make([]byte, cc.NonceSize())
	rand.Read(nc)
	ccifr := cc.Seal(nc, nc, texto, ad)
	pcc, err := cc.Open(nil, ccifr[:cc.NonceSize()], ccifr[cc.NonceSize():], ad)
	fmt.Printf("ChaCha20-Poly1305: %d bytes, %q, err=%v\n", len(ccifr), pcc, err)

	// HKDF: derivar chaves independentes de uma chave-mestra
	mestra := make([]byte, 32)
	rand.Read(mestra)
	kCPF, _ := hkdf.Key(sha256.New, mestra, nil, "campo:cpf", 32)
	kEmail, _ := hkdf.Key(sha256.New, mestra, nil, "campo:email", 32)
	kCPF2, _ := hkdf.Key(sha256.New, mestra, nil, "campo:cpf", 32)
	fmt.Println("HKDF cpf  :", hex.EncodeToString(kCPF)[:16]+"...")
	fmt.Println("HKDF email:", hex.EncodeToString(kEmail)[:16]+"...")
	fmt.Println("determinismo (mesmo contexto, mesma chave):", bytes.Equal(kCPF, kCPF2))
}
