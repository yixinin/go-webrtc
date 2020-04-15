package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var DefaultRoom *Room
var DefaultChat *Chat
var DefaultReflect *Reflect

func main() {
	var config = new(Config)
	config.Stun = []string{
		"stun:stun.voipgate.com:3478",
		"stun:stun.ideasip.com",
	}

	DefaultRoom = NewRoom(config)
	DefaultChat = NewChat()
	DefaultReflect = NewRedlect()
	go DefaultReflect.HandleReflect()

	var g = gin.Default()
	g.POST("/getAnswer", GetAnswer)
	// g.POST("/getCandidate", GetCandidate)
	g.POST("/sendCandidate", SendCandidate)
	g.GET("/Test", func(c *gin.Context) {
		c.String(200, "test ...")
	})
	g.POST("/sendSdp", SendSdp)
	g.POST("/sendCand", SendCand)
	g.POST("/pollSdp", PollSdp)
	g.POST("/pollCand", PollCandidate)

	g.POST("/reflect", ReflectF)
	g.POST("/reflectCand", ReflectCand)

	g.StaticFS("/static", http.Dir("static"))
	g.StaticFile("/index", "static/index.html")
	g.StaticFile("/index.html", "static/index.html")

	g.Run("0.0.0.0:8000")
}
