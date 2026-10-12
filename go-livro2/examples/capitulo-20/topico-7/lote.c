#include "lote.h"
#include "_cgo_export.h"

int processar(int etapas, uintptr_t h) {
	for (int i = 1; i <= etapas; i++) {
		progresso(h, i * 100 / etapas);
	}
	return etapas;
}
