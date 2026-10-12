package main

import (
	"fmt"
	"uuid"
)

func main() {
	id := uuid.New() // equivale a NewV4 (aleatório)
	fmt.Println(len(id.String()), id.String()[14] == '4')

	v7 := uuid.NewV7() // ordenável por tempo
	fmt.Println(v7.String()[14] == '7')

	lido, err := uuid.Parse("123e4567-e89b-12d3-a456-426614174000")
	fmt.Println(lido, err)
}
