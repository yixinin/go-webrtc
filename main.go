package main

import "github.com/gin-gonic/gin"

var DefaultRoom *Room

func main() {
	var config = new(Config)
	config.Stun = []string{
		"stun:stun.voipgate.com:3478",
	}

	DefaultRoom = NewRoom(config)

	var g = gin.Default()
	g.POST("/getAnswer", GetAnswer)
	g.POST("/getCandidate", GetCandidate)
	g.POST("/sendCandidate", SendCandidate)
	g.GET("/Test", func(c *gin.Context) {
		c.String(200, "test ...")
	})

	g.Run("0.0.0.0:8080")
}
