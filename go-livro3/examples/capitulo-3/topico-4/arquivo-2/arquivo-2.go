// Leitura ingênua com bufio.Scanner e buffer padrão.
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		n++
	}
	fmt.Printf("linhas lidas: %d\n", n)
	fmt.Println("erro:", sc.Err())
}
