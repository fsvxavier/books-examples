package main

import (
	"fmt"

	"github.com/google/uuid"
)

func main() {
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	fmt.Println("ID:", id.String(), "| versão do UUID:", id.Version())
}
