package texto

/*
#include <stdlib.h>
#include <string.h>

static char *repetir(const char *s, int vezes) {
	size_t n = strlen(s);
	char *r = malloc(n * vezes + 1);
	if (r == NULL) return NULL;
	for (int i = 0; i < vezes; i++) memcpy(r + i * n, s, n);
	r[n * vezes] = 0;
	return r;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

var ErrMemoria = errors.New("sem memoria")

// Repetir devolve s repetida vezes vezes, calculada pelo C.
func Repetir(s string, vezes int) (string, error) {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	r := C.repetir(cs, C.int(vezes))
	if r == nil {
		return "", ErrMemoria
	}
	defer C.free(unsafe.Pointer(r))
	return C.GoString(r), nil
}
