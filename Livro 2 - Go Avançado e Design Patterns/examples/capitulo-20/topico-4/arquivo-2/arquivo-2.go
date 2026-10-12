package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"unsafe"
)

func rssMB() float64 {
	b, _ := os.ReadFile("/proc/self/status")
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "VmRSS:") {
			var kb float64
			fmt.Sscanf(strings.TrimSpace(strings.TrimPrefix(l, "VmRSS:")), "%f", &kb)
			return kb / 1024
		}
	}
	return 0
}

func heapMB() float64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return float64(m.HeapAlloc) / (1 << 20)
}

func rodar(rotulo string, liberar bool) {
	texto := strings.Repeat("a", 1024)
	fmt.Printf("%-12s antes:  RSS %7.1f MB | heap do Go %5.2f MB\n", rotulo, rssMB(), heapMB())
	for i := 0; i < 200_000; i++ {
		cs := C.CString(texto)
		if liberar {
			C.free(unsafe.Pointer(cs))
		}
	}
	fmt.Printf("%-12s depois: RSS %7.1f MB | heap do Go %5.2f MB\n", rotulo, rssMB(), heapMB())
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "livre" {
		rodar("com free", true)
		return
	}
	rodar("sem free", false)
}
