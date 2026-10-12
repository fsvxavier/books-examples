package e03

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

var pedacos = func() []string {
	p := make([]string, 1000)
	for i := range p {
		p[i] = "0123456789"
	}
	return p
}()

var sink string

func BenchmarkMais(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		s := ""
		for _, p := range pedacos {
			s += p
		}
		sink = s
	}
}

func BenchmarkBuilder(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var sb strings.Builder
		for _, p := range pedacos {
			sb.WriteString(p)
		}
		sink = sb.String()
	}
}

func BenchmarkBuilderGrow(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var sb strings.Builder
		sb.Grow(10_000)
		for _, p := range pedacos {
			sb.WriteString(p)
		}
		sink = sb.String()
	}
}

func BenchmarkBytesBuffer(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var buf bytes.Buffer
		for _, p := range pedacos {
			buf.WriteString(p)
		}
		sink = buf.String() // copia os bytes para criar a string
	}
}

func BenchmarkJoin(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sink = strings.Join(pedacos, "")
	}
}

func BenchmarkSprintf(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		s := ""
		for _, p := range pedacos[:100] { // so 100: Sprintf com += e muito lento
			s = fmt.Sprintf("%s%s", s, p)
		}
		sink = s
	}
}

func BenchmarkAppendBytes(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		buf := make([]byte, 0, 10_000)
		for _, p := range pedacos {
			buf = append(buf, p...)
		}
		sink = string(buf) // uma copia
	}
}
