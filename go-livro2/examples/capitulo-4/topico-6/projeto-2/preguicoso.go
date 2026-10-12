//go:build !ansioso

package main

import "sync"

var pool = sync.OnceValue(abrirPool) // roda na primeira chamada
