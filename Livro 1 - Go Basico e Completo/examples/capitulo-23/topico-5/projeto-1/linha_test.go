package e06

import (
	"regexp"
	"strings"
	"testing"
)

const linha = `203.0.113.7 - - [04/Oct/2026:10:15:32 -0300] "GET /api/pedidos/42 HTTP/1.1" 200 1532`

var reGET = regexp.MustCompile(`"GET `)
var reStatus = regexp.MustCompile(`" (\d{3}) `)

var sinkB bool
var sinkS []string

func BenchmarkCompilaNoLaco(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		re := regexp.MustCompile(`" (\d{3}) `)
		sinkS = re.FindStringSubmatch(linha)
	}
}

func BenchmarkPrecompilada(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkS = reStatus.FindStringSubmatch(linha)
	}
}

func BenchmarkContainsGET(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkB = strings.Contains(linha, `"GET `)
	}
}

func BenchmarkRegexpGET(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkB = reGET.MatchString(linha)
	}
}

func BenchmarkStatusSemRegexp(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, resto, _ := strings.Cut(linha, `" `)
		sinkS = []string{resto[:3]}
	}
}

// entrada que mataria um motor com retrocesso exponencial
var patologica = regexp.MustCompile(`^(a+)+$`)
var entrada30 = strings.Repeat("a", 30) + "!"
var entrada3000 = strings.Repeat("a", 3000) + "!"

func BenchmarkPatologica30(b *testing.B) {
	for b.Loop() {
		sinkB = patologica.MatchString(entrada30)
	}
}

func BenchmarkPatologica3000(b *testing.B) {
	for b.Loop() {
		sinkB = patologica.MatchString(entrada3000)
	}
}
