package main

import (
	"database/sql"
	"database/sql/driver"
	"expvar"
	"fmt"
)

type drv struct{}

func (drv) Open(string) (driver.Conn, error) { return nil, nil }

func tentar(nome string, f func()) {
	defer func() { fmt.Printf("%-22s panic: %v\n", nome, recover()) }()
	f()
}

func main() {
	sql.Register("meu", drv{})
	tentar("sql.Register (2x)", func() { sql.Register("meu", drv{}) })
	tentar("sql.Register (nil)", func() { sql.Register("outro", nil) })

	expvar.NewInt("pedidos")
	tentar("expvar.NewInt (2x)", func() { expvar.NewInt("pedidos") })
}
