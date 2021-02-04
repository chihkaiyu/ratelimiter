package strategy

import "github.com/chihkaiyu/ratelimiter/base/ctx"

type Strategy interface {
	Acquire(context ctx.CTX, key string) (bool, int, error)
}
