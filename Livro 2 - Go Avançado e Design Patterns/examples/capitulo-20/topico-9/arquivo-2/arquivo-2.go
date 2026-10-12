package nativo

/*
#cgo nocallback vaziaPromete
#cgo noescape incrementarPromete
#cgo nocallback incrementarPromete

static int vazia(void) { return 1; }
static int vaziaPromete(void) { return 1; }

static void incrementar(int *p) { (*p)++; }
static void incrementarPromete(int *p) { (*p)++; }
*/
import "C"

func Vazia() int        { return int(C.vazia()) }
func VaziaPromete() int { return int(C.vaziaPromete()) }

func Incrementar() int {
	var x C.int
	C.incrementar(&x)
	return int(x)
}

func IncrementarPromete() int {
	var x C.int
	C.incrementarPromete(&x)
	return int(x)
}
