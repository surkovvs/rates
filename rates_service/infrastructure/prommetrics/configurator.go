package prommetrics

import (
	"database/sql"

	"go.uber.org/zap"
)

func InitPromMetrics(db *sql.DB, logger *zap.Logger) {
	GatherMetrics("rates_service", logger)
	NewDBStats(db, "postgres")
	objectives := map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	NewSummaryVec("endpoints",
		"Конечное время исполнения ручки",
		objectives,
		"handler", "response_status")
	NewSummaryVec("provider_API",
		"Конечное время выполнения запроса к внешнему API",
		objectives,
		"API", "status")
	NewSummaryVec("DB",
		"Выполнение запроса в БД",
		objectives,
		"storager", "method", "status")
}
