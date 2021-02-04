package redis

import (
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/chihkaiyu/dcard-homework/base/ctx"
)

type impl struct {
	client *redis.Client
}

func NewRedis(addr, password string) Service {
	return &impl{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
		}),
	}
}

func (im *impl) Ping(context ctx.CTX) error {
	_, err := im.client.Ping(context).Result()
	if err != nil {
		context.WithField("err", err).Error("client.Ping failed")
		return err
	}

	return nil
}

func (im *impl) RunScript(context ctx.CTX, script *redis.Script, keys []string, args ...interface{}) (interface{}, error) {
	value, err := script.Run(context, im.client, keys, args).Result()
	if err != nil && err != redis.Nil {
		context.WithField("err", err).Error("script.Run failed")
	}

	return value, err
}

func (im *impl) Get(context ctx.CTX, key string) ([]byte, error) {
	value, err := im.client.Get(context, key).Bytes()
	if err != nil {
		context.WithField("err", err).Error("client.Get failed")
		return []byte{}, err
	}

	return value, err
}

func (im *impl) Set(context ctx.CTX, key string, value []byte, ttl time.Duration) error {
	if err := im.client.Set(context, key, value, ttl).Err(); err != nil {
		context.WithField("err", err).Error("client.Set failed")
		return err
	}

	return nil
}
