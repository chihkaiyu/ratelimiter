package api

import (
	"flag"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/chihkaiyu/dcard-homework/base/ctx"
)

var (
	env = flag.String("env", "", "dev")
)

// Cors sets required headers for cors and clients
func Cors() gin.HandlerFunc {
	ch := cors.New(cors.Config{
		AllowOriginFunc: func(s string) bool {
			if *env == "dev" && strings.HasPrefix(s, "http://localhost") {
				return true
			}

			return regexp.MustCompile("https?://.*dcard.com.tw").MatchString(s)
		},
		AllowMethods: []string{"GET", "POST", "PATCH", "DELETE", "PUT"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	})

	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			ch(c)
		}

		c.Next()
	}
}

// AddContext adds context into gin
func AddContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("ctx", ctx.Background())
		c.Next()
	}
}

// SetClientIP sets client IP from request header
func SetClientIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		remoteIP := strings.Split(c.Request.RemoteAddr, ":")[0]
		c.Request.Header.Add("true-client-ip", remoteIP)
	}
}
