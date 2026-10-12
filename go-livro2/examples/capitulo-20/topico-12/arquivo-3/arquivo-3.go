package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"runtime"
)

//go:noinline
func vazar() {
	for range 3 {
		_ = C.CString("esta memoria nunca volta") // sem C.free
	}
}

func main() {
	vazar()
	runtime.GC()
	fmt.Println("fim")
}
