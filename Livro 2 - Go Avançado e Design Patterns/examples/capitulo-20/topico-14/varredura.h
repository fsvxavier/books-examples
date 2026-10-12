#ifndef VARREDURA_H
#define VARREDURA_H

#include <stddef.h>
#include <stdint.h>

// Soma os bytes de dados em 10 etapas, chamando avancar(h, percentual) ao fim
// de cada uma. Se avancar devolver diferente de zero, para e devolve -1.
long varrer(const unsigned char *dados, size_t n, uintptr_t h);

#endif
