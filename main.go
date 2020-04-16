package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var DefaultRoom *Room

func main() {
	var config = new(Config)
	config.Stun = []string{
		"stun:stun.voipgate.com:3478",
		"stun:stun.ideasip.com",
	}

	DefaultRoom = NewRoom(config)
	// go DefaultReflect.HandleReflect()
	HandleHttp()
}

func HandleHttp() {
	var g = gin.Default()
	g.StaticFS("/static", http.Dir("static"))
	g.StaticFile("/index", "static/index.html")
	g.StaticFile("/index.html", "static/index.html")
	g.StaticFile("/broadcast", "static/broadcast.html")

	g.POST("/getAnswer", SendOffer)
	g.POST("/sendCandidate", SendCandidate)

	g.POST("/sendOfferChan", SendOfferChan)
	g.POST("/sendCandChan", SendCandChan)
	g.GET("/pollCandChan", PollCandChan)

	g.Run("0.0.0.0:8000")
}
