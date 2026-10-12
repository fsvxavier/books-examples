// Package e04 compara conversões entre string e []byte: com cópia (o padrão da linguagem) e sem cópia (unsafe).
package e04

import "unsafe"

// BytesParaStringSemCopia devolve uma string que COMPARTILHA a memória de b.
// Perigoso: se b for alterado depois, a string "muda" (e as strings são imutáveis para o resto do programa).
func BytesParaStringSemCopia(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// StringParaBytesSemCopia devolve um []byte que COMPARTILHA a memória de s.
// Perigoso: o resultado NÃO PODE ser escrito (a memória de uma string pode ser somente leitura).
func StringParaBytesSemCopia(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
