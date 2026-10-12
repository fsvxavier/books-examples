package main

// #include "estatistica.h"
import "C"

import "fmt"

func Mediana(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	return float64(C.mediana((*C.double)(&v[0]), C.int(len(v))))
}

func main() {
	fmt.Println(Mediana([]float64{9, 1, 5}), Mediana([]float64{4, 1, 3, 2}))
}
