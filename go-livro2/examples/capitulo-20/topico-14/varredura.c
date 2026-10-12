#include <unistd.h>

#include "_cgo_export.h"
#include "varredura.h"

long varrer(const unsigned char *dados, size_t n, uintptr_t h) {
	long soma = 0;
	size_t passo = n / 10;
	for (int etapa = 1; etapa <= 10; etapa++) {
		size_t ini = (etapa - 1) * passo;
		size_t fim = etapa == 10 ? n : etapa * passo;
		for (size_t i = ini; i < fim; i++) soma += dados[i];
		usleep(20 * 1000); // finge trabalho demorado
		if (avancar(h, etapa * 10) != 0) return -1;
	}
	return soma;
}
