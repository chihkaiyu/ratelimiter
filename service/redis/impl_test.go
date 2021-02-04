package redis

import (
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/suite"

	"github.com/chihkaiyu/ratelimiter/base/ctx"
	"github.com/chihkaiyu/ratelimiter/base/docker"
)

var (
	mockCTX = ctx.Background()
	mockNow = time.Now()
)

type redisSuite struct {
	suite.Suite
	redis *impl
	port  string
	addr  string
}

func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(redisSuite))
}

func (s *redisSuite) SetupSuite() {
	ports, err := docker.RunExternal([]string{"redis"})
	s.NoError(err)

	s.port = ports[0]
	s.addr = "localhost:" + ports[0]
}

func (s *redisSuite) TearDownSuite() {
	s.NoError(docker.RemoveExternal())
}

func (s *redisSuite) SetupTest() {
	s.redis = NewRedis(s.addr, "").(*impl)
}

func (s *redisSuite) TearDownTest() {
	s.NoError(docker.ClearRedis(s.port))
}

func (s *redisSuite) TestPing() {
	err := s.redis.Ping(mockCTX)
	s.NoError(err)
}

func (s *redisSuite) TestGet() {
	tests := []struct {
		Desc   string
		Input  []byte
		Exp    []byte
		ExpErr error
	}{
		{
			Desc:   "get successful",
			Input:  []byte("test"),
			Exp:    []byte("test"),
			ExpErr: nil,
		},
		{
			Desc:   "key missing",
			Input:  nil,
			Exp:    nil,
			ExpErr: redis.Nil,
		},
	}

	key := "tmp"
	for _, test := range tests {
		s.SetupTest()

		if test.Input != nil {
			err := s.redis.Set(mockCTX, key, test.Input, 10*time.Minute)
			s.NoError(err, test.Desc)
		}
		act, err := s.redis.Get(mockCTX, key)
		if test.ExpErr != nil {
			s.EqualError(err, test.ExpErr.Error(), test.Desc)
		} else {
			s.Equal(test.Exp, act, test.Desc)
		}

		s.TearDownTest()
	}
}

func (s *redisSuite) TestSet() {
	err := s.redis.Set(mockCTX, "tmp", []byte("test"), 30*time.Minute)
	s.NoError(err)

	act, err := s.redis.Get(mockCTX, "tmp")
	s.NoError(err)
	s.Equal([]byte("test"), act)
}

func (s *redisSuite) TestRunScript() {
	script := `
redis.call('HSET', KEYS[1], 'timestamp', ARGV[1])
local curVal = redis.call('HINCRBY', KEYS[1], 'timestamp', ARGV[2])

return curVal`

	redisScript := redis.NewScript(script)
	key := "tmp"

	value, err := s.redis.RunScript(mockCTX, redisScript, []string{key}, mockNow.Unix(), 60)
	s.NoError(err)

	s.Equal(mockNow.Add(60*time.Second).Unix(), value.(int64))
}

func (s *redisSuite) TestIncr() {
	_, err := s.redis.client.Set(mockCTX, "tmp", []byte("10"), 30*time.Minute).Result()
	s.NoError(err)
	value, err := s.redis.Incr(mockCTX, "tmp")
	s.NoError(err)
	s.Equal(int64(11), value)

	value, err = s.redis.Incr(mockCTX, "not-exist")
	s.NoError(err)
	s.Equal(int64(1), value)
}

func (s *redisSuite) TestExpire() {
	err := s.redis.Expire(mockCTX, "tmp", 1*time.Second)
	s.NoError(err)
}

func (s *redisSuite) TestZAdd() {
	key := "tmp"
	err := s.redis.ZAdd(mockCTX, key, 5, "4")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 1, "2")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 3, "3")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 10, "5")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, -1, "1")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 999, "6")
	s.NoError(err)

	members, err := s.redis.ZRange(mockCTX, key, 0, 2)
	s.NoError(err)
	s.Len(members, 3)
	for i := 0; i < len(members); i++ {
		s.Equal(strconv.FormatInt(int64(i+1), 10), members[i])
	}

	members, err = s.redis.ZRange(mockCTX, key, 0, 10)
	s.NoError(err)
	s.Len(members, 6)
	for i := 0; i < len(members); i++ {
		s.Equal(strconv.FormatInt(int64(i+1), 10), members[i])
	}
}

func (s *redisSuite) TestZCount() {
	key := "tmp"
	err := s.redis.ZAdd(mockCTX, key, 5, "4")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 1, "2")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 3, "3")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 10, "5")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, -1, "1")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 999, "6")
	s.NoError(err)

	count, err := s.redis.ZCount(mockCTX, key, "1", "10")
	s.NoError(err)
	s.Equal(4, count)

	count, err = s.redis.ZCount(mockCTX, key, "-10", "5")
	s.NoError(err)
	s.Equal(4, count)

	count, err = s.redis.ZCount(mockCTX, key, "-inf", "inf")
	s.NoError(err)
	s.Equal(6, count)
}

func (s *redisSuite) TestZRemRangeByScore() {
	key := "tmp"
	err := s.redis.ZAdd(mockCTX, key, 5, "4")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 1, "2")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 3, "3")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 10, "5")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, -1, "1")
	s.NoError(err)
	err = s.redis.ZAdd(mockCTX, key, 999, "6")
	s.NoError(err)

	err = s.redis.ZRemRangeByScore(mockCTX, key, "1", "5")
	s.NoError(err)
	count, err := s.redis.ZCount(mockCTX, key, "-inf", "inf")
	s.NoError(err)
	s.Equal(3, count)

	err = s.redis.ZRemRangeByScore(mockCTX, key, "-10", "0")
	s.NoError(err)
	count, err = s.redis.ZCount(mockCTX, key, "-inf", "inf")
	s.NoError(err)
	s.Equal(2, count)

	err = s.redis.ZRemRangeByScore(mockCTX, key, "-inf", "inf")
	s.NoError(err)
	count, err = s.redis.ZCount(mockCTX, key, "-inf", "inf")
	s.NoError(err)
	s.Equal(0, count)
}
