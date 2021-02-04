package fixedwindow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/chihkaiyu/dcard-homework/base/ctx"
	"github.com/chihkaiyu/dcard-homework/base/docker"
	"github.com/chihkaiyu/dcard-homework/service/redis"
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

type fixedWindowSuite struct {
	suite.Suite
	redisPort   string
	fixedWindow *impl
	mockFuncs   *mockFuncs
}

func TestFixedWindowSuite(t *testing.T) {
	suite.Run(t, new(fixedWindowSuite))
}

func (s *fixedWindowSuite) SetupSuite() {
	ports, err := docker.RunExternal([]string{"redis"})
	s.NoError(err)

	s.redisPort = ports[0]
}

func (s *fixedWindowSuite) TearDownSuite() {
	s.NoError(docker.RemoveExternal())
}

func (s *fixedWindowSuite) SetupTest() {
	redis := redis.NewRedis("localhost:"+s.redisPort, "")
	s.fixedWindow = NewFixedWindow(redis).(*impl)
	*fixedWindowSize = 10
	*fixedWindowLimit = 5

	// mock functions
	s.mockFuncs = new(mockFuncs)
	timeNow = s.mockFuncs.timeNow
}

func (s *fixedWindowSuite) TearDownTest() {
	s.mockFuncs.AssertExpectations(s.T())

	s.NoError(docker.ClearRedis(s.redisPort))
}

func (s *fixedWindowSuite) TestAccquire() {
	tests := []struct {
		Desc         string
		AccquireTime []time.Time
		Exp          bool
		ExpCount     int
	}{
		{
			Desc: "normal acquire",
			AccquireTime: []time.Time{
				mockNow,
				mockNow.Add(3 * time.Second),
				mockNow.Add(6 * time.Second),
			},
			Exp:      true,
			ExpCount: 3,
		},
		{
			Desc: "normal acquire with different window",
			AccquireTime: []time.Time{
				mockNow,
				mockNow.Add(3 * time.Second),
				mockNow.Add(11 * time.Second),
			},
			Exp:      true,
			ExpCount: 1,
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
			Exp:      false,
			ExpCount: 6,
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
				mockNow.Add(10 * time.Second),
				mockNow.Add(11 * time.Second),
			},
			Exp:      true,
			ExpCount: 2,
		},
	}

	key := "localhost"
	for _, test := range tests {
		s.SetupTest()

		var act bool
		var actCount int
		var err error
		for _, t := range test.AccquireTime {
			s.mockFuncs.On("timeNow").Return(t).Once()
			act, actCount, err = s.fixedWindow.Acquire(mockCTX, key)
			s.NoError(err, test.Desc)
		}

		s.Equal(test.Exp, act, test.Desc)
		s.Equal(test.ExpCount, actCount, test.Desc)

		s.TearDownTest()
	}
}
