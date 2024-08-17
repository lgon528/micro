package redis

import (
	"context"
	"net"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type interceptor struct {
	beforeProcess         func(ctx context.Context, cmd redis.Cmder) (context.Context, error)
	afterProcess          func(ctx context.Context, cmd redis.Cmder) error
	beforeProcessPipeline func(ctx context.Context, cmds []redis.Cmder) (context.Context, error)
	afterProcessPipeline  func(ctx context.Context, cmds []redis.Cmder) error
}

func (i *interceptor) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (i *interceptor) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		var err error
		ctx, err = i.beforeProcess(ctx, cmd)
		if err != nil {
			return err
		}

		err = next(ctx, cmd)
		if err != nil {
			return err
		}

		return i.afterProcess(ctx, cmd)
	}
}

func (i *interceptor) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		var err error
		ctx, err = i.beforeProcessPipeline(ctx, cmds)
		if err != nil {
			return err
		}

		err = next(ctx, cmds)
		if err != nil {
			return err
		}

		return i.afterProcessPipeline(ctx, cmds)
	}
}

func traceInterceptor(compName string, opts *redis.Options, logger *zap.Logger) *interceptor {
	tracer := otel.Tracer(compName)
	attrs := []attribute.KeyValue{
		semconv.NetHostPortKey.String(opts.Addr),
		semconv.DBNameKey.Int(opts.DB),
		semconv.DBSystemRedis,
	}

	return newInterceptor().
		setBeforeProcess(func(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
			ctx, span := tracer.Start(ctx, cmd.FullName(), trace.WithAttributes(attrs...))
			span.SetAttributes(
				semconv.DBOperationKey.String(cmd.Name()),
				semconv.DBStatementKey.String(cast.ToString(cmd.Args())),
			)
			return ctx, nil
		}).
		setAfterProcess(func(ctx context.Context, cmd redis.Cmder) error {
			span := trace.SpanFromContext(ctx)
			if err := cmd.Err(); err != nil && err != redis.Nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "ok")
			}

			span.End()
			return nil
		}).
		setBeforeProcessPipeline(func(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
			ctx, span := tracer.Start(ctx, "pipeline", nil, trace.WithAttributes(attrs...))
			span.SetAttributes(
				semconv.DBOperationKey.String(getCmdsName(cmds)),
			)
			return ctx, nil
		}).
		setAfterProcessPipeline(func(ctx context.Context, cmds []redis.Cmder) error {
			span := trace.SpanFromContext(ctx)
			for _, cmd := range cmds {
				if err := cmd.Err(); err != nil && err != redis.Nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					span.End()
					return nil
				}
			}
			span.SetStatus(codes.Ok, "ok")
			span.End()
			return nil
		})
}

func newInterceptor() *interceptor {
	return &interceptor{
		beforeProcess: func(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
			return ctx, nil
		},
		afterProcess: func(ctx context.Context, cmd redis.Cmder) error {
			return nil
		},
		beforeProcessPipeline: func(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
			return ctx, nil
		},
		afterProcessPipeline: func(ctx context.Context, cmds []redis.Cmder) error {
			return nil
		},
	}
}

func getCmdsName(cmds []redis.Cmder) string {
	cmdNameMap := map[string]bool{}
	cmdName := []string{}
	for _, cmd := range cmds {
		if !cmdNameMap[cmd.Name()] {
			cmdName = append(cmdName, cmd.Name())
			cmdNameMap[cmd.Name()] = true
		}
	}
	return strings.Join(cmdName, "_")
}

func (i *interceptor) setBeforeProcess(p func(ctx context.Context, cmd redis.Cmder) (context.Context, error)) *interceptor {
	i.beforeProcess = p
	return i
}

func (i *interceptor) setAfterProcess(p func(ctx context.Context, cmd redis.Cmder) error) *interceptor {
	i.afterProcess = p
	return i
}

func (i *interceptor) setBeforeProcessPipeline(p func(ctx context.Context, cmds []redis.Cmder) (context.Context, error)) *interceptor {
	i.beforeProcessPipeline = p
	return i
}

func (i *interceptor) setAfterProcessPipeline(p func(ctx context.Context, cmds []redis.Cmder) error) *interceptor {
	i.afterProcessPipeline = p
	return i
}
