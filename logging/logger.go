package logging

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Logger interface {
	DebugCtx(ctx context.Context, msg string, fields ...zap.Field)
	InfoCtx(ctx context.Context, msg string, fields ...zap.Field)
	WarnCtx(ctx context.Context, msg string, fields ...zap.Field)
	ErrorCtx(ctx context.Context, msg string, fields ...zap.Field)
	DPanicCtx(ctx context.Context, msg string, fields ...zap.Field)
	PanicCtx(ctx context.Context, msg string, fields ...zap.Field)
	FatalCtx(ctx context.Context, msg string, fields ...zap.Field)

	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	DPanic(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)

	// 提供 Sugar() 方法，返回 SugaredLogger 接口
	Sugar() SugaredLogger
}

type logger struct {
	zapLogger *zap.Logger
}

func NewLogger(zapLogger *zap.Logger) Logger {
	return &logger{zapLogger: zapLogger}
}

func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.zapLogger.Debug(msg, fields...)
}

func (l *logger) Info(msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, fields...)
}

func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, fields...)
}

func (l *logger) Error(msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, fields...)
}

func (l *logger) DPanic(msg string, fields ...zap.Field) {
	l.zapLogger.DPanic(msg, fields...)
}

func (l *logger) Panic(msg string, fields ...zap.Field) {
	l.zapLogger.Panic(msg, fields...)
}

func (l *logger) Fatal(msg string, fields ...zap.Field) {
	l.zapLogger.Fatal(msg, fields...)
}

func (l *logger) DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.Debug(msg, l.withContext(ctx, fields...)...)
}

func (l *logger) InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, l.withContext(ctx, fields...)...)
}

func (l *logger) WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, l.withContext(ctx, fields...)...)
}

func (l *logger) ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, l.withContext(ctx, fields...)...)
}

func (l *logger) DPanicCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.DPanic(msg, l.withContext(ctx, fields...)...)
}

func (l *logger) PanicCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.Panic(msg, l.withContext(ctx, fields...)...)
}

func (l *logger) FatalCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.zapLogger.Fatal(msg, l.withContext(ctx, fields...)...)
}

// 根据需要实现其他方法

// withContext 从 context 中提取 traceID，并添加到日志字段中
func (l *logger) withContext(ctx context.Context, fields ...zap.Field) []zap.Field {
	if ctx == nil {
		return fields
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		fields = append(fields, zap.String("trace_id", traceID))
	}

	// 如果有其他需要从 context 中提取的字段，可以在这里处理

	return fields
}

func (l *logger) Sugar() SugaredLogger {
	return &sugaredLogger{
		zapSugaredLogger: l.zapLogger.Sugar(),
	}
}
