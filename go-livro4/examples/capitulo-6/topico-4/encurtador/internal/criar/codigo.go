package criar

import (
	"crypto/rand"

	"exemplos/livro4/capitulo-6/topico-4/encurtador/internal/dominio"
)

// CodigoAleatorio é o gerador de produção. O alfabeto tem 32 símbolos e 256 é
// múltiplo de 32, então o resto da divisão de um byte sorteado não tem viés.
func CodigoAleatorio() string {
	var b [dominio.TamanhoCodigo]byte
	_, _ = rand.Read(b[:])
	for i := range b {
		b[i] = dominio.Alfabeto[int(b[i])%len(dominio.Alfabeto)]
	}
	return string(b[:])
}
