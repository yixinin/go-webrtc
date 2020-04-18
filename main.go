package main

import (
	"go-webrtc/config"
	"go-webrtc/protocol"

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
}
