package main

import (
	"go-webrtc/config"
	"go-webrtc/room"
)

type Server struct {
	rooms  map[int32]*room.Room
	config *config.Config
}

func NewServer(conf *config.Config) *Server {
	return &Server{
		rooms:  make(map[int32]*room.Room, 30),
		config: conf,
	}
}
