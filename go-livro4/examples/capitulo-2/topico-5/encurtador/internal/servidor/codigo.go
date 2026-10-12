package servidor

import "crypto/rand"

// O alfabeto tem 32 símbolos, sem caracteres confundíveis (0/o, 1/l/i).
// Como 256 é múltiplo de 32, usar o resto da divisão de um byte aleatório
// por 32 não introduz viés: todos os símbolos são igualmente prováveis.
const (
	alfabeto       = "abcdefghjkmnpqrstuvwxyz23456789"
	tamanhoCodigo  = 6
	tentativasMax  = 5
	alfabetoTamano = len(alfabeto)
)

// geradorCodigo produz um código novo a cada chamada.
type geradorCodigo func() string

func codigoAleatorio() string {
	var b [tamanhoCodigo]byte
	// rand.Read nunca devolve erro desde o Go 1.24: em caso de falha, o
	// programa termina de forma irrecuperável.
	_, _ = rand.Read(b[:])
	for i := range b {
		b[i] = alfabeto[int(b[i])%alfabetoTamano]
	}
	return string(b[:])
}
