package slidingwindow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/chihkaiyu/ratelimiter/base/ctx"
	"github.com/chihkaiyu/ratelimiter/base/docker"
	"github.com/chihkaiyu/ratelimiter/service/redis"
)

var (
	mockCTX = ctx.Background()
	mockNow = time.Date(2021, time.February, 1, 0, 0, 0, 0, time.UTC)
)

type mockFuncs struct {
	mock.Mock
}

func (m *mockFuncs) timeNow() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

type slidingWindowSuite struct {
	suite.Suite
	redisPort     string
	slidingWindow *impl
	mockFuncs     *mockFuncs
}

func TestSlidingWindowSuite(t *testing.T) {
	suite.Run(t, new(slidingWindowSuite))
}

func (s *slidingWindowSuite) SetupSuite() {
	ports, err := docker.RunExternal([]string{"redis"})
	s.NoError(err)

	s.redisPort = ports[0]
}

func (s *slidingWindowSuite) TearDownSuite() {
	s.NoError(docker.RemoveExternal())
}

func (s *slidingWindowSuite) SetupTest() {
	redis := redis.NewRedis("localhost:"+s.redisPort, "")
	s.slidingWindow = NewSlidingWindow(redis).(*impl)
	*slidingWindowSize = 10
	*slidingWindowLimit = 5

	// mock functions
	s.mockFuncs = new(mockFuncs)
	timeNow = s.mockFuncs.timeNow
}

func (s *slidingWindowSuite) TearDownTest() {
	s.mockFuncs.AssertExpectations(s.T())

	s.NoError(docker.ClearRedis(s.redisPort))
}

func (s *slidingWindowSuite) TestAccquire() {
	tests := []struct {
		Desc         string
		AccquireTime []time.Time
		Exp          []bool
		ExpCount     []int
	}{
		{
			Desc: "normal acquire",
			AccquireTime: []time.Time{
				mockNow,
				mockNow.Add(3 * time.Second),
				mockNow.Add(6 * time.Second),
			},
			Exp:      []bool{true, true, true},
			ExpCount: []int{1, 2, 3},
		},
		{
			Desc: "normal acquire with different window",
			AccquireTime: []time.Time{
				mockNow,
				mockNow.Add(3 * time.Second),
				mockNow.Add(11 * time.Second),
			},
			Exp:      []bool{true, true, true},
			ExpCount: []int{1, 2, 2},
		},
		{
			Desc: "acquire failed",
			AccquireTime: []time.Time{
				mockNow,
				mockNow.Add(1 * time.Second),
				mockNow.Add(2 * time.Second),
				mockNow.Add(3 * time.Second),
				mockNow.Add(4 * time.Second),
				mockNow.Add(5 * time.Second),
			},
			Exp:      []bool{true, true, true, true, true, false},
			ExpCount: []int{1, 2, 3, 4, 5, 5},
		},
		{
			Desc: "acquire failed but success at different window",
			AccquireTime: []time.Time{
				mockNow,
				mockNow.Add(1 * time.Second),
				mockNow.Add(2 * time.Second),
				mockNow.Add(3 * time.Second),
				mockNow.Add(4 * time.Second),
				mockNow.Add(5 * time.Second),
				mockNow.Add(13 * time.Second),
				mockNow.Add(14 * time.Second),
			},
			Exp:      []bool{true, true, true, true, true, false, true, true},
			ExpCount: []int{1, 2, 3, 4, 5, 5, 3, 3},
		},
	}

	key := "localhost"
	for _, test := range tests {
		s.SetupTest()

		s.Len(test.Exp, len(test.AccquireTime))
		s.Len(test.ExpCount, len(test.AccquireTime))

		var act bool
		var actCount int
		var err error
		for i, t := range test.AccquireTime {
			s.mockFuncs.On("timeNow").Return(t).Once()
			act, actCount, err = s.slidingWindow.Acquire(mockCTX, key)
			s.NoError(err, test.Desc)
			s.Equal(test.Exp[i], act, test.Desc)

			s.Equal(test.ExpCount[i], actCount, test.Desc)
		}

		s.TearDownTest()
	}

}
