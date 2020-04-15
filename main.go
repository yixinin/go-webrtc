package main

import "github.com/gin-gonic/gin"

var DefaultRoom *Room
var DefaultChat *Chat

func main() {
	var config = new(Config)
	config.Stun = []string{
		"stun:stun.voipgate.com:3478",
		"stun:stun.ideasip.com",
	}

	DefaultRoom = NewRoom(config)
	DefaultChat = NewChat()
	var g = gin.Default()
	g.POST("/getAnswer", GetAnswer)
	g.POST("/getCandidate", GetCandidate)
	g.POST("/sendCandidate", SendCandidate)
	g.GET("/Test", func(c *gin.Context) {
		c.String(200, "test ...")
	})
	g.POST("/sendSdp", SendSdp)
	g.POST("/sendCand", SendCand)
	g.POST("/pollSdp", PollSdp)
	g.POST("/pollCand", PollCandidate)

	g.POST("/reflect", ReflectF)

	g.Run("0.0.0.0:8000")
}
