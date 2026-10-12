//go:build ansioso

package main

var poolAnsioso = abrirPool() // roda na carga do pacote, sempre

func pool() string { return poolAnsioso }
