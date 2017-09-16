package api

import "github.com/gin-gonic/gin"

func CreateRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "success",
		})
	})
	return r
}
