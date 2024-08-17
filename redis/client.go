package redis

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewClient(opts *redis.Options, logger *zap.Logger) *redis.Client {
	client := redis.NewClient(opts)
	client.AddHook(traceInterceptor("redis", opts, logger))
	return client
}
