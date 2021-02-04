package tokenbucket

import (
	"flag"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/chihkaiyu/ratelimiter/base/ctx"
	"github.com/chihkaiyu/ratelimiter/service/ratelimiter/strategy"
	"github.com/chihkaiyu/ratelimiter/service/redis"
)

// local remain = math.min(newSize, tonumber(ARGV[4]))
// newSize = math.max(remain - 1, 0)

const (
	// ARGV: nowTimestamp, nowNanoSecond, refillPerSecond, bucketSize
	script = `
local newSize = tonumber(ARGV[4])
local oldData = redis.call('HMGET', KEYS[1], 'ts', 'tsNano', 'tokens')
if oldData[1] then
	local secDiff = tonumber(ARGV[1]) - tonumber(oldData[1])
	local nanosecDiff = tonumber(ARGV[2]) - tonumber(oldData[2])
	newSize = tonumber(oldData[3]) + tonumber(ARGV[3]) * (secDiff + nanosecDiff / 1000000000)
end

local remain = math.min(math.floor(newSize), tonumber(ARGV[4]))
if newSize >= 1 then
	newSize = newSize - 1
end

redis.call('HMSET', KEYS[1], 'ts', ARGV[1], 'tsNano', ARGV[2], 'tokens', newSize) 
redis.call('EXPIRE', KEYS[1], math.ceil(tonumber(ARGV[4]) / tonumber(ARGV[3])))

return remain
`
)

var (
	timeNow = time.Now

	bucketSize      = flag.Int("bucketsize", 60, "token bucket size")
	refillPerSecond = flag.Float64("refill_per_second", 1, "token buckect refill speed (in second)")
)

type impl struct {
	redis       redis.Service
	redisScript *goredis.Script
	size        int
	refill      float64
}

func NewTokenBucket(
	redis redis.Service,
) strategy.Strategy {
	return &impl{
		redis:       redis,
		redisScript: goredis.NewScript(script),
		size:        *bucketSize,
		refill:      *refillPerSecond,
	}
}

func (im *impl) Acquire(context ctx.CTX, key string) (bool, int, error) {
	now := timeNow()
	nano := now.Nanosecond()

	redisKey := fmt.Sprintf("tokenbucket:%s", key)
	remain, err := im.redis.RunScript(
		context,
		im.redisScript,
		[]string{redisKey},
		now.Unix(),
		nano,
		*refillPerSecond,
		*bucketSize,
	)
	if err != nil {
		context.WithField("err", err).Error("redis.RunScript failed")
		return false, 0, err
	}

	if remain.(int64) <= 0 {
		return false, *bucketSize, nil
	}

	return true, *bucketSize - int(remain.(int64)) + 1, nil
}
