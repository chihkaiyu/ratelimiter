package tokenbucket

import (
	"flag"
	"fmt"

	"github.com/chihkaiyu/dcard-homework/base/ctx"
	"github.com/chihkaiyu/dcard-homework/service/ratelimiter/strategy"
	"github.com/chihkaiyu/dcard-homework/service/redis"
)

const (
	script = `
local newSize = 0
local oldData = redis.call('HGET', KEYS[1], 'timestamp', 'token')
if oldData[1] then
	newSize = 

newSize = math.max(tonumber(oldData[3]) - tonumber(ARGV[3]) * (secDiff + nanosecDiff / 1000000000), 0)

return curVal
`
)

var (
	bucketSize      = flag.Int("bucketsize", 60, "token bucket size")
	refillPerSecond = flag.Int("refill_per_second", 1, "token buckect refill speed (in second)")
)

type impl struct {
	redis  redis.Service
	size   int
	refill int
}

func NewTokenBucket(
	redis redis.Service,
) strategy.Strategy {
	return &impl{
		redis:  redis,
		size:   *bucketSize,
		refill: *refillPerSecond,
	}
}

func (im *impl) Acquire(context ctx.CTX, key string) (bool, int, error) {
	return false, 0, fmt.Errorf("not implement yet")
}
