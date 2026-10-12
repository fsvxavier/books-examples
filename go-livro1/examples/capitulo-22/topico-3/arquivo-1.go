package main

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
)

// Parametros do Argon2id (minimo recomendado pela OWASP: 19 MiB, t=2, p=1).
type Params struct {
	MemoriaKiB uint32
	Iteracoes  uint32
	Paralelo   uint8
	TamSal     uint32
	TamChave   uint32
}

var Padrao = Params{MemoriaKiB: 19 * 1024, Iteracoes: 2, Paralelo: 1, TamSal: 16, TamChave: 32}

var ErrFormato = errors.New("hash em formato invalido")

// Hash devolve uma string no formato PHC: tudo o que e preciso para verificar
// (algoritmo, parametros, sal e hash) fica junto.
func Hash(senha string, p Params) (string, error) {
	sal := make([]byte, p.TamSal)
	if _, err := rand.Read(sal); err != nil {
		return "", err
	}
	chave := argon2.IDKey([]byte(senha), sal, p.Iteracoes, p.MemoriaKiB, p.Paralelo, p.TamChave)
	b64 := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.MemoriaKiB, p.Iteracoes, p.Paralelo, b64.EncodeToString(sal), b64.EncodeToString(chave)), nil
}

// Verificar recalcula com os parametros guardados e compara em tempo constante.
// Devolve tambem se o hash esta desatualizado e deve ser refeito.
func Verificar(senha, codificado string, atual Params) (ok, refazer bool, err error) {
	partes := strings.Split(codificado, "$")
	if len(partes) != 6 || partes[1] != "argon2id" {
		return false, false, ErrFormato
	}
	var versao int
	var p Params
	if _, err := fmt.Sscanf(partes[2], "v=%d", &versao); err != nil || versao != argon2.Version {
		return false, false, ErrFormato
	}
	if _, err := fmt.Sscanf(partes[3], "m=%d,t=%d,p=%d", &p.MemoriaKiB, &p.Iteracoes, &p.Paralelo); err != nil {
		return false, false, ErrFormato
	}
	b64 := base64.RawStdEncoding
	sal, err := b64.DecodeString(partes[4])
	if err != nil {
		return false, false, ErrFormato
	}
	esperado, err := b64.DecodeString(partes[5])
	if err != nil {
		return false, false, ErrFormato
	}
	calc := argon2.IDKey([]byte(senha), sal, p.Iteracoes, p.MemoriaKiB, p.Paralelo, uint32(len(esperado)))
	ok = subtle.ConstantTimeCompare(calc, esperado) == 1
	refazer = p.MemoriaKiB < atual.MemoriaKiB || p.Iteracoes < atual.Iteracoes
	return ok, refazer, nil
}

func medir(nome string, f func()) {
	ini := time.Now()
	f()
	fmt.Printf("%-34s %s\n", nome, time.Since(ini).Round(time.Millisecond))
}

func main() {
	senha := "correct horse battery staple"
	sal := make([]byte, 16)
	rand.Read(sal)

	// SHA-256 simples: rapido demais para guardar senhas
	medir("sha256 x 1.000.000 (uma CPU)", func() {
		for range 1_000_000 {
			sha256.Sum256([]byte(senha))
		}
	})
	medir("pbkdf2-sha256 600.000 iteracoes", func() { pbkdf2.Key(sha256.New, senha, sal, 600_000, 32) })
	medir("scrypt N=2^17 r=8 p=1", func() { scrypt.Key([]byte(senha), sal, 1<<17, 8, 1, 32) })
	for _, custo := range []int{10, 12, 14} {
		medir(fmt.Sprintf("bcrypt custo %d", custo), func() { bcrypt.GenerateFromPassword([]byte(senha), custo) })
	}
	medir("argon2id 19MiB t=2 p=1 (OWASP)", func() { argon2.IDKey([]byte(senha), sal, 2, 19*1024, 1, 32) })
	medir("argon2id 64MiB t=3 p=4", func() { argon2.IDKey([]byte(senha), sal, 3, 64*1024, 4, 32) })

	fmt.Println()
	h1, _ := Hash(senha, Padrao)
	h2, _ := Hash(senha, Padrao)
	fmt.Println(h1)
	fmt.Println("mesma senha, outro sal, hash diferente:", h1 != h2)

	ok, refazer, err := Verificar(senha, h1, Padrao)
	fmt.Println("senha certa:  ", ok, refazer, err)
	ok, _, _ = Verificar("senha errada", h1, Padrao)
	fmt.Println("senha errada: ", ok)

	// hash antigo, com parametros mais fracos: valida, mas deve ser refeito no login
	fraco := Params{MemoriaKiB: 8 * 1024, Iteracoes: 1, Paralelo: 1, TamSal: 16, TamChave: 32}
	hf, _ := Hash(senha, fraco)
	ok, refazer, _ = Verificar(senha, hf, Padrao)
	fmt.Println("hash fraco:   ", ok, "refazer:", refazer)

	// bcrypt ignora (ou rejeita) senhas acima de 72 bytes
	longa := strings.Repeat("a", 73)
	_, err = bcrypt.GenerateFromPassword([]byte(longa), 10)
	fmt.Println("bcrypt 73 bytes:", err)
}
