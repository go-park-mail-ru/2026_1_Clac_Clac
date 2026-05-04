package tracer

import (
	"context"
	"time"

	dbMetric "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/metrics/postgres"
	"github.com/jackc/pgx/v5"
)

type ctxKey string

const startTimeKey ctxKey = "db_start_time"

type PrometheusTracer struct{}

func (pt *PrometheusTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, startTimeKey, time.Now())
}

func (pt *PrometheusTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	startTime, ok := ctx.Value(startTimeKey).(time.Time)
	if !ok {
		return
	}

	duration := time.Since(startTime).Seconds()

	status := "success"
	if data.Err != nil {
		status = "error"
	}

	dbMetric.DbQueryDuration.WithLabelValues("sql_query").Observe(duration)
	dbMetric.DbQueriesTotal.WithLabelValues(status).Inc()
}
