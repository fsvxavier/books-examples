package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func memoria() int64 {
	runtime.GC()
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	return int64(s.HeapAlloc)
}

func tamanhoDaRaspagem(reg *prometheus.Registry) (series int, bytesTexto int) {
	mfs, _ := reg.Gather()
	var buf bytes.Buffer
	enc := expfmt.NewEncoder(&buf, expfmt.NewFormat(expfmt.TypeTextPlain))
	for _, mf := range mfs {
		series += len(mf.Metric)
		enc.Encode(mf)
	}
	return series, buf.Len()
}

func main() {
	// 5 rotas x 4 status: cardinalidade limitada e conhecida
	{
		reg := prometheus.NewRegistry()
		c := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "req_total", Help: "x"}, []string{"rota", "status"})
		reg.MustRegister(c)
		for r := range 5 {
			for _, st := range []string{"200", "400", "404", "500"} {
				c.WithLabelValues("rota"+strconv.Itoa(r), st).Inc()
			}
		}
		s, b := tamanhoDaRaspagem(reg)
		fmt.Printf("rota x status:      series=%-7d raspagem=%d bytes\n", s, b)
	}
	// um rotulo com ID de usuario: uma serie por usuario
	for _, n := range []int{1000, 100000} {
		reg := prometheus.NewRegistry()
		c := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "req_total", Help: "x"}, []string{"usuario"})
		reg.MustRegister(c)
		antes := memoria()
		for u := range n {
			c.WithLabelValues("usuario-" + strconv.Itoa(u)).Inc()
		}
		depois := memoria()
		s, b := tamanhoDaRaspagem(reg)
		fmt.Printf("por usuario %-6d series=%-7d raspagem=%-8d bytes memoria~%d KiB\n", n, s, b, (depois-antes)/1024)
	}
}
