package main

import (
	"fmt"
	"testing"

	"github.com/go-acme/lego/v3/log"
	"github.com/pion/webrtc/v2"
)

func TestOffer(t *testing.T) {
	var m = webrtc.MediaEngine{}
	var api = webrtc.NewAPI(webrtc.WithMediaEngine(m))
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.ideasip.com"},
			},
		},
	}
	var conn, err = api.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		log.Println(err)
		return
	}
	offer, err := conn.CreateOffer(nil)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(offer.SDP)
}
