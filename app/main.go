package main

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/siruspen/logrus"

	"github.com/chihkaiyu/dcard-homework/api"
)

var (
	port = flag.Int("port", 9000, "api server port")
)

func main() {
	flag.Parse()

	router := gin.Default()
	router.Use(api.Cors())

	rg := router.Group("/api/v1")
	// TODO: add rate limiter here
	rg.Use(
		api.AddContext(), api.SetClientIP(),
	)

	if err := router.Run(fmt.Sprintf(":%d", *port)); err != nil {
		logrus.Panicf("router.Run failed, err: %v", err)
	}
}
