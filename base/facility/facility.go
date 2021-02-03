package facility

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/braintree/manners"
	"github.com/gin-gonic/gin"
	"github.com/siruspen/logrus"
)

const (
	defaultDebugAddr = ":8080"
)

var (
	shutdownHandlers     = map[int][]handlerFunc{}
	shutdownHandlerMutex = sync.Mutex{}
)

type (
	handlerFunc func() error

	serveOption struct {
		addr          string
		ginRouter     *gin.Engine
		debugAddr     string
		shutdownLevel int
	}

	// Option is an alias for functional argument in ServeHTTP
	Option func(*serveOption)
)

func addShutdownHandler(handler handlerFunc, options ...Option) {
	o := &serveOption{}
	for _, opt := range options {
		opt(o)
	}

	shutdownHandlerMutex.Lock()
	defer shutdownHandlerMutex.Unlock()
	shutdownHandlers[o.shutdownLevel] = append(shutdownHandlers[o.shutdownLevel], handler)
}

func hookShutdownHandler(name string) {
	term := make(chan os.Signal)
	termCheck := make(chan os.Signal)
	allSignal := make(chan os.Signal)
	// seems like it's impossible to get SIGKILL
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	signal.Notify(termCheck, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	signal.Notify(allSignal)

	type orderedHandler struct {
		level    int
		handlers []handlerFunc
	}

	// log all signals
	go func() {
		for s := range allSignal {
			// filter out the useless signal "child existed"
			if s == syscall.SIGCHLD {
				continue
			}
			logrus.Info("Got signal: " + s.String())
		}
	}()

	go func() {
		<-term
		logrus.Info("Got terminated signal")
		termTime := time.Now()

		oHlrs := []orderedHandler{}
		for lvl, hlrs := range shutdownHandlers {
			oHlrs = append(oHlrs, orderedHandler{level: lvl, handlers: hlrs})
		}
		// sort in ascending order
		sort.Slice(oHlrs, func(i, j int) bool { return oHlrs[i].level < oHlrs[j].level })

		var termWg sync.WaitGroup
		for _, oHlr := range oHlrs {
			// Registered callbacks at the same level run concurrently
			for i := range oHlr.handlers {
				cb := oHlr.handlers[i]
				termWg.Add(1)
				go func() {
					defer termWg.Done()
					cb()
				}()
			}
			termWg.Wait()
		}
		t := float64(time.Since(termTime) / time.Second)
		logrus.Info(fmt.Sprintf("shutdown callbacks finished in %fs", t))
	}()

	go func() {
		<-termCheck
		time.Sleep(25 * time.Second)
		logrus.Warn("Time limit of graceful shutdown exceeded.")
	}()
}

func startMonitorServer(addr string) {
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			logrus.Warnf("Cannot start the monitor server: %v", err)
		}
	}()
}

func Serve(name string, opts ...Option) error {
	o := &serveOption{}
	for _, opt := range opts {
		opt(o)
	}

	if o.debugAddr == "" {
		o.debugAddr = defaultDebugAddr
	}

	hookShutdownHandler(name)
	startMonitorServer(o.debugAddr)

	gin.EnableJsonDecoderUseNumber()
	addShutdownHandler(func() error {
		manners.Close()
		return nil
	}, WithShutdownLevel(o.shutdownLevel))

	return manners.ListenAndServe(o.addr, o.ginRouter)
}

// WithShutdownLevel adds shutdown level to shutdown handler
func WithShutdownLevel(level int) Option {
	return func(o *serveOption) {
		o.shutdownLevel = level
	}
}

// WithGinRouter adds a gin router
func WithGinRouter(addr string, router *gin.Engine) Option {
	return func(o *serveOption) {
		o.addr = addr
		o.ginRouter = router
	}
}
