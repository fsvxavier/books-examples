package main

/*
extern void gancho(int n);

static int rodar(int n) {
	for (int i = 0; i < n; i++) {
		gancho(i);
	}
	return n;
}

typedef struct { int id; } Item;
extern Item *obter(void);
static int usar(void) {
	Item *p = obter();
	return p->id;
}
*/
import "C"

import (
	"errors"
	"fmt"
)

type valor struct{ id C.int }

//export gancho
func gancho(n C.int) {
	if n == 2 {
		panic(errors.New("deu ruim no passo 2"))
	}
	fmt.Println("gancho", n)
}

//export obter
func obter() *C.Item {
	v := new(C.Item)
	v.id = 7
	return v
}

func main() {
	func() {
		defer func() { fmt.Println("recuperado no Go:", recover()) }()
		C.rodar(5)
	}()
	fmt.Println("o programa continua")

	func() {
		defer func() { fmt.Println("recuperado:", recover()) }()
		fmt.Println(C.usar())
	}()
}
