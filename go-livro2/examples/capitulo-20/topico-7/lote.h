#ifndef LOTE_H
#define LOTE_H

#include <stdint.h>

// Chama progresso(h, percentual) a cada etapa; devolve quantas etapas rodou.
int processar(int etapas, uintptr_t h);

#endif
