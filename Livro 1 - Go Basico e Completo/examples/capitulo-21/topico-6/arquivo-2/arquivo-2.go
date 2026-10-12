package main

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// DBStats expoe as estatisticas do pool do database/sql (Capitulo 20) como
// metricas. Um Collector le o valor NO MOMENTO da raspagem, sem goroutine propria.
type DBStats struct {
	db                      *sql.DB
	abertas, emUso, ociosas *prometheus.Desc
	esperas, esperaSegundos *prometheus.Desc
}

func NovoDBStats(db *sql.DB, nome string) *DBStats {
	lbl := prometheus.Labels{"banco": nome}
	d := func(n, ajuda string) *prometheus.Desc {
		return prometheus.NewDesc("db_pool_"+n, ajuda, nil, lbl)
	}
	return &DBStats{
		db:             db,
		abertas:        d("open_connections", "Conexoes abertas (em uso + ociosas)."),
		emUso:          d("in_use_connections", "Conexoes em uso."),
		ociosas:        d("idle_connections", "Conexoes ociosas."),
		esperas:        d("wait_total", "Vezes que alguem esperou por uma conexao."),
		esperaSegundos: d("wait_seconds_total", "Tempo total esperando por conexao."),
	}
}

func (c *DBStats) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range []*prometheus.Desc{c.abertas, c.emUso, c.ociosas, c.esperas, c.esperaSegundos} {
		ch <- d
	}
}

func (c *DBStats) Collect(ch chan<- prometheus.Metric) {
	s := c.db.Stats()
	ch <- prometheus.MustNewConstMetric(c.abertas, prometheus.GaugeValue, float64(s.OpenConnections))
	ch <- prometheus.MustNewConstMetric(c.emUso, prometheus.GaugeValue, float64(s.InUse))
	ch <- prometheus.MustNewConstMetric(c.ociosas, prometheus.GaugeValue, float64(s.Idle))
	ch <- prometheus.MustNewConstMetric(c.esperas, prometheus.CounterValue, float64(s.WaitCount))
	ch <- prometheus.MustNewConstMetric(c.esperaSegundos, prometheus.CounterValue, s.WaitDuration.Seconds())
}

func main() {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT 1").WillReturnRows(sqlmock.NewRows([]string{"x"}).AddRow(1))
	var x int
	db.QueryRow("SELECT 1").Scan(&x) // abre uma conexao

	reg := prometheus.NewRegistry()
	reg.MustRegister(NovoDBStats(db, "principal"))

	esperado := `
# HELP db_pool_idle_connections Conexoes ociosas.
# TYPE db_pool_idle_connections gauge
db_pool_idle_connections{banco="principal"} 1
# HELP db_pool_in_use_connections Conexoes em uso.
# TYPE db_pool_in_use_connections gauge
db_pool_in_use_connections{banco="principal"} 0
# HELP db_pool_open_connections Conexoes abertas (em uso + ociosas).
# TYPE db_pool_open_connections gauge
db_pool_open_connections{banco="principal"} 1
`
	err := testutil.GatherAndCompare(reg, strings.NewReader(esperado),
		"db_pool_idle_connections", "db_pool_in_use_connections", "db_pool_open_connections")
	fmt.Println("conferencia com testutil:", err)

	// quantas series o coletor do Go publica?
	reg2 := prometheus.NewRegistry()
	reg2.MustRegister(collectors.NewGoCollector())
	mfs, _ := reg2.Gather()
	n := 0
	for _, mf := range mfs {
		n += len(mf.Metric)
	}
	fmt.Println("familias de metricas do GoCollector:", len(mfs), "series:", n)
	for _, mf := range mfs {
		if strings.HasPrefix(mf.GetName(), "go_goroutines") || strings.HasPrefix(mf.GetName(), "go_gc_duration") || mf.GetName() == "go_memstats_heap_alloc_bytes" {
			fmt.Println("  ", mf.GetName(), "-", mf.GetType())
		}
	}
}
