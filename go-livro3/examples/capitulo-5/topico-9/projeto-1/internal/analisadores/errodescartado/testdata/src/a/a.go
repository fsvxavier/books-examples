package a

import (
	"fmt"
	"os"
	"strconv"
)

func exemplo() {
	os.Remove("x")            // want `o erro devolvido por os.Remove é ignorado`
	n, _ := strconv.Atoi("1") // want `o erro devolvido por strconv.Atoi é descartado com _`
	fmt.Println(n)
	if err := os.Remove("y"); err != nil {
		fmt.Println(err)
	}
}
