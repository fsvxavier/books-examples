package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

func main() {
	g, ctx := errgroup.WithContext(context.Background())
	for n := range Gerar(ctx, g, 2, []int{1, 2, 3, 4, 5}) {
		fmt.Println("item", n)
	}
	fmt.Println("fim:", g.Wait())
}
