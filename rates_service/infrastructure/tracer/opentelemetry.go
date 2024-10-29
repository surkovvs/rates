package tracer

import (
	"context"
	"rates_service/config"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/stats"
)

var exporter *otlptrace.Exporter
var enabled bool

func InitTracer(cfg config.AppCfg) stats.Handler {
	enabled = cfg.Trace.Enable
	if !enabled {
		return nil
	}
	exporter = otlptracegrpc.NewUnstarted(
		// otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(cfg.Trace.Host+":"+cfg.Trace.Port),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: time.Second * 2,
			MaxInterval:     time.Second * 3,
			MaxElapsedTime:  time.Second * 5,
		}),
	)
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	otel.SetLogger(NewStdoutLogger())
	return otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(provider))
}

func Start(ctx context.Context) error {
	if enabled {
		return exporter.Start(ctx)
	}
	return nil
}
func Shutdown(ctx context.Context) error {
	if enabled {
		return exporter.Shutdown(ctx)
	}
	return nil
}
