package redis

import (
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/chihkaiyu/dcard-homework/base/ctx"
)

type Service interface {
	Ping(context ctx.CTX) error

	RunScript(context ctx.CTX, script *redis.Script, keys []string, args ...interface{}) (interface{}, error)

	Get(context ctx.CTX, key string) ([]byte, error)

	Set(context ctx.CTX, key string, value []byte, ttl time.Duration) error
}
