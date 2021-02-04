package fixedwindow

import (
	"flag"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/chihkaiyu/ratelimiter/base/ctx"
	"github.com/chihkaiyu/ratelimiter/service/ratelimiter/strategy"
	"github.com/chihkaiyu/ratelimiter/service/redis"
)

var (
	timeNow = time.Now

	fixedWindowSize  = flag.Int("fixed_window_size", 60, "fixed window size (in second)")
	fixedWindowLimit = flag.Int("fixed_window_limit", 60, "fixed window limit")
)

type impl struct {
	redis  redis.Service
	size   int
	litmit int
}

func NewFixedWindow(
	redis redis.Service,
) strategy.Strategy {
	return &impl{
		redis:  redis,
		size:   *fixedWindowSize,
		litmit: *fixedWindowLimit,
	}
}

func (im *impl) Acquire(context ctx.CTX, key string) (bool, int, error) {
	now := timeNow()
	window := now.Unix() / int64(im.size)
	redisKey := fmt.Sprintf("fixed_window:%s:%d", key, window)
	value, err := im.redis.Incr(context, redisKey)
	if err != nil && err != redis.Nil {
		context.WithFields(logrus.Fields{
			"err": err,
			"key": redisKey,
		}).Error("redis.Incr failed")
		return false, 0, err
	}
	defer func() {
		if err := im.redis.Expire(context, redisKey, time.Duration(im.size)*time.Second); err != nil {
			context.WithFields(logrus.Fields{
				"err": err,
				"key": redisKey,
			}).Error("redis.Expire failed")
		}
	}()

	if value > int64(im.litmit) {
		return false, int(value), nil
	}

	return true, int(value), nil
}
