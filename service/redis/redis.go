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

	// ZAdd adds member score to sorted set of given key
	ZAdd(context ctx.CTX, key string, score int, member string) error

	// ZCount counts the member whose score are between given min and max score
	ZCount(context ctx.CTX, key string, min, max string) (int, error)

	ZRange(context ctx.CTX, key string, start, end int) ([]string, error)

	// ZRemRangeByScore removes the member whose scores are between given min and max
	ZRemRangeByScore(context ctx.CTX, key string, min, max string) error
}
