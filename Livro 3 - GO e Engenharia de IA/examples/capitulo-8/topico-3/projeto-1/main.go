package main

import (
	"fmt"
	"os"
)

func main() {
	if err := memoria(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}
