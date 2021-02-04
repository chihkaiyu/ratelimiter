package redis

import (
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/chihkaiyu/ratelimiter/base/ctx"
)

type impl struct {
	client *redis.Client
}

func NewRedis(addr, password string) Service {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	_, err := client.Ping(ctx.Background()).Result()
	if err != nil {
		panic(err)
	}
	return &impl{
		client: client,
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

func (im *impl) Incr(context ctx.CTX, key string) (int64, error) {
	value, err := im.client.Incr(context, key).Result()
	if err != nil {
		context.WithField("err", err).Error("client.Incr failed")
		return 0, err
	}

	return value, nil
}

func (im *impl) Expire(context ctx.CTX, key string, ttl time.Duration) error {
	_, err := im.client.Expire(context, key, ttl).Result()
	if err != nil {
		context.WithField("err", err).Error("client.Expire failed")
		return err
	}

	return nil
}

func (im *impl) ZAdd(context ctx.CTX, key string, score int, member string) error {
	_, err := im.client.ZAdd(context, key, &redis.Z{
		Score:  float64(score),
		Member: member,
	}).Result()
	if err != nil {
		context.WithField("err", err).Error("client.ZAdd failed")
		return err
	}

	return nil
}

func (im *impl) ZRange(context ctx.CTX, key string, start, end int) ([]string, error) {
	members, err := im.client.ZRange(context, key, int64(start), int64(end)).Result()
	if err != nil {
		context.WithField("err", err).Error("client.ZRange failed")
		return []string{}, err
	}

	return members, nil
}

func (im *impl) ZCount(context ctx.CTX, key string, min, max string) (int, error) {
	count, err := im.client.ZCount(context, key, min, max).Result()
	if err != nil {
		context.WithField("err", err).Error("client.ZCount failed")
		return 0, err
	}

	return int(count), nil
}

func (im *impl) ZRemRangeByScore(context ctx.CTX, key string, min, max string) error {
	_, err := im.client.ZRemRangeByScore(context, key, min, max).Result()
	if err != nil {
		context.WithField("err", err).Error("client.ZRemRangeByScore failed")
		return err
	}
	return nil
}
