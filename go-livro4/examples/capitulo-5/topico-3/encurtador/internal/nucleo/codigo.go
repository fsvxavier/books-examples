package nucleo

import (
	"crypto/rand"

	"exemplos/livro4/capitulo-5/topico-3/encurtador/internal/nucleo/dominio"
)

// CodigoAleatorio é o Gerador de produção. O alfabeto tem 32 símbolos e 256 é
// múltiplo de 32, então o resto da divisão de um byte sorteado não tem viés.
// rand.Read nunca devolve erro: se o sistema não fornecer entropia, o programa
// termina de forma irrecuperável.
func CodigoAleatorio() string {
	var b [dominio.TamanhoCodigo]byte
	_, _ = rand.Read(b[:])
	for i := range b {
		b[i] = dominio.Alfabeto[int(b[i])%len(dominio.Alfabeto)]
	}
	return string(b[:])
}
