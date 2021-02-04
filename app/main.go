package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/chihkaiyu/ratelimiter/api"
	"github.com/chihkaiyu/ratelimiter/service/ratelimiter"
	"github.com/chihkaiyu/ratelimiter/service/redis"
)

var (
	port      = flag.Int("port", 9000, "api server port")
	redisAddr = flag.String("redis_addr", "localhost:6379", "redis addr: host:port")
)

func main() {
	flag.Parse()

	redis := redis.NewRedis(*redisAddr, "")
	limiter := ratelimiter.NewRateLimiter(redis)
	ratelimiter := api.NewRateLimiter(limiter, gin.H{"error": "too many request"}, http.StatusTooManyRequests)

	router := gin.Default()
	router.Use(api.Cors())
	rg := router.Group("/api/v1")
	// TODO: add rate limiter here
	rg.Use(
		api.AddContext(), api.SetClientIP(), ratelimiter.Acquire(),
	)
	rg.GET("/ping", func(c *gin.Context) {
		api.JSON(c, http.StatusOK)
	})

	if err := router.Run(fmt.Sprintf(":%d", *port)); err != nil {
		logrus.Panicf("router.Run failed, err: %v", err)
	}
}
