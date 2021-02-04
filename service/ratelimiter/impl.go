package ratelimiter

import (
	"flag"

	"github.com/chihkaiyu/dcard-homework/base/ctx"
	"github.com/chihkaiyu/dcard-homework/service/ratelimiter/strategy"
	"github.com/chihkaiyu/dcard-homework/service/ratelimiter/strategy/fixedwindow"
	"github.com/chihkaiyu/dcard-homework/service/ratelimiter/strategy/tokenbucket"
	"github.com/chihkaiyu/dcard-homework/service/redis"
)

var (
	rateLimiterStrategy = flag.String("ratelimiter_strategy", "fixedwindow", "strategy for rate limiting")
)

type impl struct {
	strategy strategy.Strategy
}

func NewRateLimiter(
	redis redis.Service,
) Service {
	var stra strategy.Strategy
	switch *rateLimiterStrategy {
	case "tokenbucket":
		stra = tokenbucket.NewTokenBucket(redis)
	case "fixedwindow":
		stra = fixedwindow.NewFixedWindow(redis)
	default:
		stra = fixedwindow.NewFixedWindow(redis)
	}
	return &impl{
		strategy: stra,
	}
}

func (im *impl) AcquireByIP(context ctx.CTX, ip string) (bool, int, error) {
	permit, count, err := im.strategy.Acquire(context, ip)
	if err != nil {
		context.WithField("err", err).Error("strategy.acquire failed")
		return false, 0, err
	}

	return permit, count, nil
}
