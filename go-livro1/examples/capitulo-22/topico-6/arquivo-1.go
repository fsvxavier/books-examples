package main

import (
	"bytes"
	"crypto"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/mldsa"
	"crypto/mlkem"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
)

func main() {
	msg := []byte("contrato v3 - aprovado")

	// Ed25519: assinatura moderna, chaves e assinaturas pequenas
	pubEd, privEd, _ := ed25519.GenerateKey(rand.Reader)
	sigEd := ed25519.Sign(privEd, msg)
	fmt.Printf("Ed25519   publica=%dB assinatura=%dB valida=%v adulterada=%v\n",
		len(pubEd), len(sigEd), ed25519.Verify(pubEd, msg, sigEd), ed25519.Verify(pubEd, append(msg, '!'), sigEd))

	// ECDSA P-256: o padrao do TLS e dos certificados
	privEc, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	h := sha256.Sum256(msg)
	sigEc, _ := ecdsa.SignASN1(rand.Reader, privEc, h[:])
	fmt.Printf("ECDSA P256 assinatura~%dB valida=%v\n", len(sigEc), ecdsa.VerifyASN1(&privEc.PublicKey, h[:], sigEc))

	// RSA 2048: assinatura PSS e cifra OAEP
	privRSA, _ := rsa.GenerateKey(rand.Reader, 2048)
	sigRSA, _ := rsa.SignPSS(rand.Reader, privRSA, crypto.SHA256, h[:], nil)
	errRSA := rsa.VerifyPSS(&privRSA.PublicKey, crypto.SHA256, h[:], sigRSA, nil)
	fmt.Printf("RSA-2048  assinatura=%dB valida=%v\n", len(sigRSA), errRSA == nil)
	seg := []byte("chave-simetrica-de-32-bytes-....")
	cifrada, _ := rsa.EncryptOAEP(sha256.New(), rand.Reader, &privRSA.PublicKey, seg, nil)
	plana, _ := rsa.DecryptOAEP(sha256.New(), nil, privRSA, cifrada, nil)
	fmt.Printf("RSA-OAEP  cifrado=%dB roundtrip=%v (limite de entrada: %d bytes)\n",
		len(cifrada), bytes.Equal(seg, plana), privRSA.PublicKey.Size()-2*sha256.Size-2)
	_, err := rsa.GenerateKey(rand.Reader, 512)
	fmt.Println("RSA-512:  ", err)

	// X25519: acordo de chaves (Diffie-Hellman)
	a, _ := ecdh.X25519().GenerateKey(rand.Reader)
	b, _ := ecdh.X25519().GenerateKey(rand.Reader)
	sa, _ := a.ECDH(b.PublicKey())
	sb, _ := b.ECDH(a.PublicKey())
	fmt.Printf("X25519    segredo compartilhado=%dB igual nos dois lados=%v\n", len(sa), bytes.Equal(sa, sb))

	// ML-KEM-768 (pos-quantico): encapsulamento de chave
	dk, _ := mlkem.GenerateKey768()
	ek := dk.EncapsulationKey()
	chaveEnv, ct := ek.Encapsulate()
	chaveRec, _ := dk.Decapsulate(ct)
	fmt.Printf("ML-KEM-768 publica=%dB cifrado=%dB chave=%dB igual=%v\n",
		len(ek.Bytes()), len(ct), len(chaveEnv), bytes.Equal(chaveEnv, chaveRec))

	// ML-DSA-65 (pos-quantico, Go 1.27): assinatura
	sk, _ := mldsa.GenerateKey(mldsa.MLDSA65())
	sigMl, _ := sk.Sign(nil, msg, nil)
	okMl := mldsa.Verify(sk.PublicKey(), msg, sigMl, nil)
	fmt.Printf("ML-DSA-65 publica=%dB assinatura=%dB valida=%v\n", len(sk.PublicKey().Bytes()), len(sigMl), okMl == nil)

	// PEM / PKCS#8: como chaves sao guardadas em arquivo
	der, _ := x509.MarshalPKCS8PrivateKey(privEd)
	bloco := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	linhas := strings.Split(strings.TrimSpace(string(bloco)), "\n")
	fmt.Println(linhas[0], "...", linhas[len(linhas)-1])
	lido, _ := pem.Decode(bloco)
	k, _ := x509.ParsePKCS8PrivateKey(lido.Bytes)
	fmt.Println("chave lida do PEM igual a original:", k.(ed25519.PrivateKey).Equal(privEd))
}
