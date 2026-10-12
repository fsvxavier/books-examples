package main

/*
#include <sys/syscall.h>
#include <unistd.h>

static int na_thread_principal(void) {
	return syscall(SYS_gettid) == getpid();
}
*/
import "C"

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "travar" {
		// no init, a goroutine principal fica presa a thread principal do processo
		runtime.LockOSThread()
	}
}

func main() {
	mudancas := 0
	anterior := C.na_thread_principal() != 0
	for range 300 {
		time.Sleep(time.Millisecond)
		agora := C.na_thread_principal() != 0
		if agora != anterior {
			mudancas++
		}
		anterior = agora
	}
	fmt.Printf("args=%v  na thread principal ao final: %v  mudancas de thread vistas: %d\n", os.Args[1:], anterior, mudancas)
}
