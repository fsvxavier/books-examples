package config

import (
	"os"
	"sync"
)

type Config struct{ Ambiente string }

// Carregar nao guarda nada: recebe a fonte e devolve o valor.
func Carregar(getenv func(string) string) Config {
	return Config{Ambiente: getenv("APP_ENV")}
}

// Padrao continua existindo para quem nao recebe a configuracao por parametro,
// mas agora e so uma casca fina sobre Carregar.
var Padrao = sync.OnceValue(func() Config { return Carregar(os.Getenv) })
