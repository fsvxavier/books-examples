// Command e12 usa os padrões 2, 4 e 5 de unsafe.Pointer, cada um da forma válida (Linux, x86-64).
package main

import (
	"fmt"
	"reflect"
	"syscall"
	"unsafe"
)

func main() {
	x := 42

	// Padrão 2: Pointer -> uintptr só para mostrar o endereço (e nunca converter de volta).
	fmt.Printf("padrão 2: %%p = %p | uintptr = %#x (iguais: %v)\n", &x, uintptr(unsafe.Pointer(&x)), fmt.Sprintf("%p", &x) == fmt.Sprintf("%#x", uintptr(unsafe.Pointer(&x))))

	// Padrão 4: a conversão para uintptr aparece na própria lista de argumentos da chamada de sistema.
	mensagem := []byte("padrão 4: escrito por syscall.Syscall(SYS_WRITE)\n")
	n, _, errno := syscall.Syscall(syscall.SYS_WRITE, 1, uintptr(unsafe.Pointer(&mensagem[0])), uintptr(len(mensagem)))
	fmt.Println("bytes escritos:", n, "| errno:", errno)

	// Padrão 5: o resultado de reflect.Value.Pointer é convertido NA MESMA EXPRESSÃO.
	p := (*int)(unsafe.Pointer(reflect.ValueOf(&x).Pointer()))
	*p = 7
	fmt.Println("padrão 5: x depois de escrever pelo ponteiro:", x)
}
