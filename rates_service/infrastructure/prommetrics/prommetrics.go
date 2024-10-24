package prommetrics

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type promRegistry struct {
	namespace string
	reg       *prometheus.Registry
	sumVecs   map[string]*prometheus.SummaryVec
	histoVecs map[string]*prometheus.HistogramVec
	log       *zap.Logger
}

var (
	gatherFlag bool
	promReg    promRegistry
)

func GatherMetrics(Namespace string, log *zap.Logger) {
	gatherFlag = true
	promReg = promRegistry{
		namespace: Namespace,
		reg:       prometheus.NewRegistry(),
		sumVecs:   make(map[string]*prometheus.SummaryVec),
		log:       log,
	}
	if err := promReg.reg.Register(
		collectors.NewProcessCollector(
			collectors.ProcessCollectorOpts{
				Namespace:    promReg.namespace,
				ReportErrors: true,
			},
		),
	); err != nil {
		if promReg.log != nil {
			promReg.log.Error("prommetrics", zap.Error(err))
		}
	}
}

func GetHandler() http.Handler {
	if gatherFlag {
		return promhttp.HandlerFor(promReg.reg, promhttp.HandlerOpts{Registry: promReg.reg})
	}
	return http.NewServeMux()
}

func NewDBStats(db *sql.DB, dbName string) {
	if gatherFlag {
		if err := promReg.reg.Register(collectors.NewDBStatsCollector(db, dbName)); err != nil {
			if promReg.log != nil {
				promReg.log.Error("prommetrics", zap.Error(err))
			}
		}
	}
}

func NewSummaryVec(Name, Help string, Objectives map[float64]float64, lableNames ...string) {
	if gatherFlag {
		sumVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:  promReg.namespace,
			Name:       Name,
			Help:       Help,
			Objectives: Objectives,
		},
			lableNames)
		promReg.sumVecs[Name] = sumVec
		if err := promReg.reg.Register(sumVec); err != nil {
			if promReg.log != nil {
				promReg.log.Error("prommetrics", zap.Error(err))
			}
		}
	}
}
func ObcerveSummaryVec(name string, lables ...string) func() {
	if !gatherFlag {
		return func() {}
	}
	t := time.Now()
	return func() {
		if sumVec, ok := promReg.sumVecs[name]; ok {
			sumVec.WithLabelValues(lables...).Observe(time.Since(t).Seconds())
		}
	}
}
func ObcerveSummaryVecSplit(name string) func(lables ...string) {
	if !gatherFlag {
		return func(lables ...string) {}
	}
	t := time.Now()
	return func(lables ...string) {
		if sumVec, ok := promReg.sumVecs[name]; ok {
			sumVec.WithLabelValues(lables...).Observe(time.Since(t).Seconds())
		}
	}
}

func NewHistogramVec(Name, Help string, Buckets []float64, lableNames ...string) {
	if gatherFlag {
		histo := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   promReg.namespace,
			Name:        Name,
			Help:        Help,
			ConstLabels: map[string]string{},
			Buckets:     []float64{},
		},
			lableNames)
		promReg.histoVecs[Name] = histo
		if err := promReg.reg.Register(histo); err != nil {
			if promReg.log != nil {
				promReg.log.Error("prommetrics", zap.Error(err))
			}
		}
	}
}
func ObcerveHistogramVecSplit(name string) func(lables ...string) {
	if !gatherFlag {
		return func(lables ...string) {}
	}
	t := time.Now()
	return func(lables ...string) {
		if histoVec, ok := promReg.histoVecs[name]; ok {
			histoVec.WithLabelValues(lables...).Observe(time.Since(t).Seconds())
		}
	}
}
