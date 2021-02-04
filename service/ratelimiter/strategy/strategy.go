package strategy

import "github.com/chihkaiyu/dcard-homework/base/ctx"

type Strategy interface {
	Acquire(context ctx.CTX, key string) (bool, int, error)
}
