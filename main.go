package main

import (
	"go-webrtc/config"
	"go-webrtc/protocol"
	"go-webrtc/room"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	var config = new(config.Config)
	config.Stun = []string{
		"stun:stun.voipgate.com:3478",
		"stun:stun.ideasip.com",
	}

	var srv = grpc.NewServer()
	var server = NewServer(config)
	protocol.RegisterRoomServiceServer(srv, server)
	HandleHttp()
}

func HandleHttp() {
	var g = gin.Default()
	g.StaticFS("/static", http.Dir("static"))
	g.StaticFile("/index", "static/index.html")
	g.StaticFile("/index.html", "static/index.html")
	g.StaticFile("/broadcast", "static/broadcast.html")

	g.POST("/getAnswer", room.SendOffer)
	g.POST("/sendCandidate", room.SendCandidate)

	g.POST("/sendOfferChan", room.SendOfferChan)
	g.POST("/sendCandChan", room.SendCandChan)
	g.GET("/pollCandChan", room.PollCandChan)

	g.Run("0.0.0.0:8000")
}
