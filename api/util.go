package api

import "github.com/gin-gonic/gin"

// JSON wraps gin context's JSON method and removes private field.
func JSON(c *gin.Context, code int) {
	setAllowOrigin(c)
	// context := c.MustGet("ctx").(ctx.CTX)
	count := c.MustGet("reqCount").(int)

	c.JSON(code, gin.H{
		"current_request_count": count,
	})
}

func setAllowOrigin(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	c.Header("Access-Control-Allow-Origin", origin)
}
