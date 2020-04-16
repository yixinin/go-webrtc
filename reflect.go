package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
)

// var CandidateChan = make(chan CandiateModel)

type Reflect struct {
	candidates    []*ReflectCandidate
	answerSdp     string
	addCandidates []*CandiateModel
}

type ReflectCandidate struct {
	candidate *CandiateModel
	added     bool
}

func NewRedlect() *Reflect {
	return &Reflect{
		candidates:    make([]*ReflectCandidate, 0),
		addCandidates: make([]*CandiateModel, 0),
	}
}
func (r *Reflect) HandleReflect() {
	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

	// Wait for the offer to be pasted
	sdp := <-offerChan

	offer := webrtc.SessionDescription{
		SDP:  sdp,
		Type: webrtc.SDPTypeOffer,
	}

	// We make our own mediaEngine so we can place the sender's codecs in it. Since we are echoing their RTP packet
	// back to them we are actually codec agnostic - we can accept all their codecs. This also ensures that we use the
	// dynamic media type from the sender in our answer.
	mediaEngine := webrtc.MediaEngine{}

	// Add codecs to the mediaEngine. Note that even though we are only going to echo back the sender's video we also
	// add audio codecs. This is because createAnswer will create an audioTransceiver and associated SDP and we currently
	// cannot tell it not to. The audio SDP must match the sender's codecs too...
	err := mediaEngine.PopulateFromSDP(offer)
	if err != nil {
		panic(err)
	}

	mediaCodecs := mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeVideo)
	if len(mediaCodecs) == 0 {
		mediaCodecs = mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeAudio)
		// panic("Offer contained no video codecs")
	}
	if len(mediaCodecs) == 0 {
		return
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.ideasip.com", "stun:stun.voipgate.com:3478"},
			},
		},
		SDPSemantics: webrtc.SDPSemanticsPlanB,
	}
	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	// Create Track that we send video back to browser on
	outputTrack, err := peerConnection.NewTrack(mediaCodecs[0].PayloadType, rand.Uint32(), "video", "pion")
	if err != nil {
		panic(err)
	}

	// Add this newly created track to the PeerConnection
	if _, err = peerConnection.AddTrack(outputTrack); err != nil {
		panic(err)
	}

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	// Set a handler for when a new remote track starts, this handler copies inbound RTP packets,
	// replaces the SSRC and sends them back
	peerConnection.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This is a temporary fix until we implement incoming RTCP events, then we would push a PLI only when a viewer requests it
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: track.SSRC()}})
				if errSend != nil {
					fmt.Println(errSend)
				}
			}
		}()

		fmt.Printf("Track has started, of type %d: %s \n", track.PayloadType(), track.Codec().Name)
		for {
			// Read RTP packets being sent to Pion
			rtp, readErr := track.ReadRTP()
			if readErr != nil {
				panic(readErr)
			}

			// Replace the SSRC with the SSRC of the outbound track.
			// The only change we are making replacing the SSRC, the RTP packets are unchanged otherwise
			rtp.SSRC = outputTrack.SSRC()

			if writeErr := outputTrack.WriteRTP(rtp); writeErr != nil {
				panic(writeErr)
			}
		}
	})
	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})
	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			var cand = c.ToJSON()
			var add = &CandiateModel{
				Candidate: cand.Candidate,
			}
			if cand.SDPMid != nil {
				add.SdpMid = *cand.SDPMid
			}
			if cand.SDPMLineIndex != nil {
				add.SdpMlineindex = *cand.SDPMLineIndex
			}
			r.addCandidates = append(r.addCandidates, add)
		}
	})

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Output the answer in base64 so we can paste it in browser
	answerChan <- answer.SDP
	r.answerSdp = answer.SDP
	log.Println("sended answer")
	for {
		if peerConnection.ICEConnectionState() == webrtc.ICEConnectionStateCompleted {
			return
		}
		for _, v := range r.candidates {
			peerConnection.AddICECandidate(webrtc.ICECandidateInit{
				Candidate:     v.candidate.Candidate,
				SDPMLineIndex: &v.candidate.SdpMlineindex,
				SDPMid:        &v.candidate.SdpMid,
			})
		}

	}
	// Block forever
	// select {}
}
