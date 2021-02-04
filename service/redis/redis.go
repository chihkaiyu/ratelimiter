package redis

import (
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/chihkaiyu/dcard-homework/base/ctx"
)

var (
	Nil = redis.Nil
)

type Service interface {
	// Ping pings the redis server, return error when failed
	Ping(context ctx.CTX) error

	// RunScript runs lua script
	RunScript(context ctx.CTX, script *redis.Script, keys []string, args ...interface{}) (interface{}, error)

	// Get gets the result of given key
	Get(context ctx.CTX, key string) ([]byte, error)

	// Set sets the value of given key with TTL
	Set(context ctx.CTX, key string, value []byte, ttl time.Duration) error

	// Incr increases by one of given key
	Incr(context ctx.CTX, key string) (int64, error)

	// Expire sets the TTL of given key
	Expire(context ctx.CTX, key string, ttl time.Duration) error
}
