package redis

import (
	"github.com/redis/go-redis/v9"
)

func NewClient(opts *redis.Options) *redis.Client {
	client := redis.NewClient(opts)
	client.AddHook(traceInterceptor("redis", opts))
	return client
}
