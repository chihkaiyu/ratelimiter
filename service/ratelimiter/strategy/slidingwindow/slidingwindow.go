package slidingwindow

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/chihkaiyu/ratelimiter/base/ctx"
	"github.com/chihkaiyu/ratelimiter/service/ratelimiter/strategy"
	"github.com/chihkaiyu/ratelimiter/service/redis"
)

var (
	timeNow = time.Now

	slidingWindowSize  = flag.Int("sliding_window_size", 60, "sliding window size (in second)")
	slidingWindowLimit = flag.Int("sliding_window_limit", 60, "sliding window limit")
)

type impl struct {
	redis redis.Service
	size  int
	limit int
}

func NewSlidingWindow(
	redis redis.Service,
) strategy.Strategy {
	return &impl{
		redis: redis,
		size:  *slidingWindowSize,
		limit: *slidingWindowLimit,
	}
}

func (im *impl) Acquire(context ctx.CTX, key string) (bool, int, error) {
	now := timeNow()
	from := now.Add(time.Duration(-im.size) * time.Second)
	min := strconv.FormatInt(from.UnixNano(), 10)
	max := strconv.FormatInt(now.UnixNano(), 10)

	redisKey := fmt.Sprintf("sliding_window:%s", key)
	count, err := im.redis.ZCount(context, redisKey, min, max)
	if err != nil && err != redis.Nil {
		context.WithFields(logrus.Fields{
			"err": err,
			"key": redisKey,
		}).Error("redis.ZCount failed")
		return false, 0, err
	}
	defer func() {
		// we don't need the window if request doesn't appear in size seconds
		// since every time we count the number of request between now-size and now
		// there won't be any records if no request appears
		if err := im.redis.Expire(context, redisKey, time.Duration(im.size)*time.Second); err != nil {
			context.WithFields(logrus.Fields{
				"err": err,
				"key": redisKey,
			}).Error("redis.Expire failed")
		}

		// clear the records in outdated windows for reducing the data
		if err := im.redis.ZRemRangeByScore(context, redisKey, "-inf", min); err != nil {
			context.WithFields(logrus.Fields{
				"err": err,
				"key": redisKey,
			}).Error("redis.ZRemRangeByScore failed")
		}
	}()

	if count >= im.limit {
		return false, count, nil
	}

	if err := im.redis.ZAdd(context, redisKey, int(now.UnixNano()), max); err != nil {
		context.WithFields(logrus.Fields{
			"err": err,
			"key": redisKey,
		}).Error("redis.ZAdd failed")
		return false, count, err
	}

	return true, count + 1, nil
}
