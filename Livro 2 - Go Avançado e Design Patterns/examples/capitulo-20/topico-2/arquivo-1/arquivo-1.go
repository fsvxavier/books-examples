package main

/*
#include <stdio.h>
#include <stdlib.h>

static void dizer(const char *msg) {
	printf("[C] %s\n", msg);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func main() {
	msg := C.CString("olá, vindo do Go")
	defer C.free(unsafe.Pointer(msg))
	C.dizer(msg)
	fmt.Println("[Go] fim")
}
