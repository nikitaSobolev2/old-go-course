package observability // инициализация OpenTelemetry для демо (stdout traces)

import (
	"context" // контекст для Shutdown TracerProvider

	"go.opentelemetry.io/otel"                              // глобальный API SetTracerProvider
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace" // экспорт спанов в stdout (не для prod)
	"go.opentelemetry.io/otel/propagation"                  // W3C TraceContext + Baggage
	sdktrace "go.opentelemetry.io/otel/sdk/trace"           // реализация TracerProvider с batcher
)

// InitTracer sets up a stdout trace exporter (replace with OTLP in production).
func InitTracer() (func(context.Context) error, error) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint()) // человекочитаемый JSON в консоль
	if err != nil {
		return nil, err // не удалось создать exporter
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp)) // батчинг экспорта спанов
	otel.SetTracerProvider(tp) // регистрируем глобально для otelhttp и ручных спанов
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator( // распространение trace-id между сервисами
		propagation.TraceContext{}, propagation.Baggage{}))
	return tp.Shutdown, nil // вызывающий обязан вызвать shutdown при завершении процесса
}
