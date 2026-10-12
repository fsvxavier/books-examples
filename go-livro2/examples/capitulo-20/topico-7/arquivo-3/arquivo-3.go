package main

/*
#cgo LDFLAGS: -lpthread
#include <pthread.h>

extern void aviso(int id, int passo);

static void *rotina(void *arg) {
	int id = (int)(long)arg;
	for (int passo = 1; passo <= 3; passo++) {
		aviso(id, passo);
	}
	return NULL;
}

// cria n threads do SO, que o runtime do Go nao conhece, e espera todas
static void disparar(int n) {
	pthread_t t[8];
	for (int i = 0; i < n; i++) pthread_create(&t[i], NULL, rotina, (void *)(long)(i + 1));
	for (int i = 0; i < n; i++) pthread_join(t[i], NULL);
}
*/
import "C"

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu     sync.Mutex
	avisos = map[int][]int{}
)

//export aviso
func aviso(id, passo C.int) {
	mu.Lock()
	defer mu.Unlock()
	avisos[int(id)] = append(avisos[int(id)], int(passo))
}

func main() {
	C.disparar(4)

	ids := make([]int, 0, len(avisos))
	for id := range avisos {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		fmt.Printf("thread do C %d chamou o Go com os passos %v\n", id, avisos[id])
	}
}
