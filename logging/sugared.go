package logging

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type SugaredLogger interface {
	DebugCtx(ctx context.Context, args ...interface{})
	InfoCtx(ctx context.Context, args ...interface{})
	WarnCtx(ctx context.Context, args ...interface{})
	ErrorCtx(ctx context.Context, args ...interface{})
	DebugfCtx(ctx context.Context, template string, args ...interface{})
	InfofCtx(ctx context.Context, template string, args ...interface{})
	WarnfCtx(ctx context.Context, template string, args ...interface{})
	ErrorfCtx(ctx context.Context, template string, args ...interface{})
	FatalfCtx(ctx context.Context, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(args ...interface{})

	Desugar() Logger
}

type sugaredLogger struct {
	zapSugaredLogger *zap.SugaredLogger
}

func NewSugaredLogger(zapLogger *zap.Logger) SugaredLogger {
	return &sugaredLogger{zapSugaredLogger: zapLogger.Sugar()}
}

func (l *sugaredLogger) DebugCtx(ctx context.Context, args ...interface{}) {
	args = l.withContext(ctx, args)
	l.zapSugaredLogger.Debug(args...)
}

func (l *sugaredLogger) InfoCtx(ctx context.Context, args ...interface{}) {
	args = l.withContext(ctx, args)
	l.zapSugaredLogger.Info(args...)
}

func (l *sugaredLogger) WarnCtx(ctx context.Context, args ...interface{}) {
	args = l.withContext(ctx, args)
	l.zapSugaredLogger.Warn(args...)
}

func (l *sugaredLogger) ErrorCtx(ctx context.Context, args ...interface{}) {
	args = l.withContext(ctx, args)
	l.zapSugaredLogger.Error(args...)
}

func (l *sugaredLogger) FatalfCtx(ctx context.Context, args ...interface{}) {
	args = l.withContext(ctx, args)
	l.zapSugaredLogger.Error(args...)
}

func (l *sugaredLogger) DebugfCtx(ctx context.Context, template string, args ...interface{}) {
	template, args = l.withContextf(ctx, template, args)
	l.zapSugaredLogger.Debugf(template, args...)
}

func (l *sugaredLogger) InfofCtx(ctx context.Context, template string, args ...interface{}) {
	template, args = l.withContextf(ctx, template, args)
	l.zapSugaredLogger.Infof(template, args...)
}

func (l *sugaredLogger) WarnfCtx(ctx context.Context, template string, args ...interface{}) {
	template, args = l.withContextf(ctx, template, args)
	l.zapSugaredLogger.Warnf(template, args...)
}

func (l *sugaredLogger) ErrorfCtx(ctx context.Context, template string, args ...interface{}) {
	template, args = l.withContextf(ctx, template, args)
	l.zapSugaredLogger.Errorf(template, args...)
}

func (l *sugaredLogger) Debug(args ...interface{}) {
	l.zapSugaredLogger.Debug(args...)
}

func (l *sugaredLogger) Info(args ...interface{}) {
	l.zapSugaredLogger.Info(args...)
}

func (l *sugaredLogger) Warn(args ...interface{}) {
	l.zapSugaredLogger.Warn(args...)
}

func (l *sugaredLogger) Error(args ...interface{}) {
	l.zapSugaredLogger.Error(args...)
}

func (l *sugaredLogger) Debugf(template string, args ...interface{}) {
	l.zapSugaredLogger.Debugf(template, args...)
}

func (l *sugaredLogger) Infof(template string, args ...interface{}) {
	l.zapSugaredLogger.Infof(template, args...)
}

func (l *sugaredLogger) Warnf(template string, args ...interface{}) {
	l.zapSugaredLogger.Warnf(template, args...)
}

func (l *sugaredLogger) Errorf(template string, args ...interface{}) {
	l.zapSugaredLogger.Errorf(template, args...)
}

func (l *sugaredLogger) Fatalf(args ...interface{}) {
	l.zapSugaredLogger.Error(args...)
}

func (l *sugaredLogger) Desugar() Logger {
	return &logger{
		zapLogger: l.zapSugaredLogger.Desugar(),
	}
}

// 辅助函数，用于在 args 中添加 trace_id
func (l *sugaredLogger) withContext(ctx context.Context, args []interface{}) []interface{} {
	if ctx == nil {
		return args
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		// 将 trace_id 添加到 args 中
		args = append(args, "trace_id", traceID)
	}

	return args
}

// 辅助函数，用于在模板和 args 中添加 trace_id
func (l *sugaredLogger) withContextf(ctx context.Context, template string, args []interface{}) (string, []interface{}) {
	if ctx == nil {
		return template, args
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		// 修改模板和 args，将 trace_id 添加到日志中
		template = template + " | trace_id: %s"
		args = append(args, traceID)
	}

	return template, args
}
