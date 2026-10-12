package main

/*
static int ler(int *p) { return *p; }
*/
import "C"

import "fmt"

func main() {
	defer func() { fmt.Println("recover:", recover()) }()
	fmt.Println("antes")
	fmt.Println(C.ler(nil))
	fmt.Println("depois")
}
