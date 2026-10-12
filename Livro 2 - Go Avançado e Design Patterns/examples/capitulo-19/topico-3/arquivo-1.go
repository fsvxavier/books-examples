// Command e02 usa unsafe.Pointer no padrão 1 (reinterpretar *T1 como *T2) e mostra o que dá errado.
package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

// Cabecalho tem layout fixo: sem preenchimento (4+2+2 bytes, bem alinhado).
type Cabecalho struct {
	Magico uint32
	Versao uint16
	Flags  uint16
}

func main() {
	// 1. Reinterpretar float64 como uint64: é o que math.Float64bits faz.
	f := 3.14159
	bits := *(*uint64)(unsafe.Pointer(&f))
	fmt.Printf("bits via unsafe:        %#016x\n", bits)
	fmt.Printf("math.Float64bits:       %#016x (igual: %v)\n", math.Float64bits(f), bits == math.Float64bits(f))

	// 2. Ver os bytes de uma struct: depende da ordem dos bytes (endianness) da máquina.
	h := Cabecalho{Magico: 0xCAFEBABE, Versao: 1, Flags: 0x0203}
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(&h)), unsafe.Sizeof(h))
	fmt.Printf("\nbytes da struct (ordem da máquina): % x\n", bytes)

	// 3. O jeito portável: encoding/binary, com a ordem escolhida.
	buf := make([]byte, 8)
	binary.BigEndian.PutUint32(buf[0:], h.Magico)
	binary.BigEndian.PutUint16(buf[4:], h.Versao)
	binary.BigEndian.PutUint16(buf[6:], h.Flags)
	fmt.Printf("bytes via encoding/binary (big-endian, igual em qualquer máquina): % x\n", buf)
}
