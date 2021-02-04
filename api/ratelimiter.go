package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/chihkaiyu/ratelimiter/base/ctx"
	"github.com/chihkaiyu/ratelimiter/service/ratelimiter"
)

type RateLimiter struct {
	errorBody interface{}
	errorCode int
	limiter   ratelimiter.Service
}

func NewRateLimiter(limiter ratelimiter.Service, errorBody interface{}, errorCode int) *RateLimiter {
	return &RateLimiter{
		limiter:   limiter,
		errorBody: errorBody,
		errorCode: errorCode,
	}
}

func (rl *RateLimiter) Acquire() gin.HandlerFunc {
	return func(c *gin.Context) {
		context := c.MustGet("ctx").(ctx.CTX)
		ip := c.GetHeader("true-client-ip")

		permit, count, err := rl.limiter.AcquireByIP(context, ip)
		if err != nil {
			context.WithFields(logrus.Fields{
				"err": err,
				"ip":  ip,
			}).Error("limiter.AcquireByIP failed")

			setAllowOrigin(c)
			c.JSON(rl.errorCode, rl.errorBody)
			c.Abort()
			return
		}

		if !permit {
			setAllowOrigin(c)
			c.JSON(rl.errorCode, rl.errorBody)
			c.Abort()
			return
		}

		c.Set("reqCount", count)
		c.Next()
	}
}
