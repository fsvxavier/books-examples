package config

import (
	"os"
	"sync"
)

type Config struct{ Ambiente string }

var (
	once      sync.Once
	instancia Config
)

// Obter le APP_ENV UMA vez e guarda para sempre.
func Obter() Config {
	once.Do(func() {
		instancia = Config{Ambiente: os.Getenv("APP_ENV")}
	})
	return instancia
}
