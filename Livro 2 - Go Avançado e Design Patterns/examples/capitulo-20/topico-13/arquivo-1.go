package main

import (
	"fmt"

	"github.com/ebitengine/purego"
)

func main() {
	lib, err := purego.Dlopen("libm.so.6", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}
	var raiz func(float64) float64
	purego.RegisterLibFunc(&raiz, lib, "sqrt")
	fmt.Println("sqrt(2) pela libm, sem cgo:", raiz(2))

	var pot func(float64, float64) float64
	purego.RegisterLibFunc(&pot, lib, "pow")
	fmt.Println("pow(2, 10):", pot(2, 10))
}
