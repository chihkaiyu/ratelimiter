package ratelimiter

import (
	"github.com/chihkaiyu/ratelimiter/base/ctx"
)

type Service interface {
	// AccquireByIP accquires the permission from rate limiter
	AcquireByIP(context ctx.CTX, ip string) (bool, int, error)
}
