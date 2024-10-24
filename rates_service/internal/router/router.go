package router

import (
	"rates_service/infrastructure/prommetrics"

	"github.com/go-chi/chi"
)

func InitChi() *chi.Mux {
	mux := chi.NewMux()
	mux.Mount("/debug", Profiler(nil))
	mux.Mount("/metrics", prommetrics.GetHandler())
	return mux
}
