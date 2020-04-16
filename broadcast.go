package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-acme/lego/v3/log"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var candChan = make(chan *CandiateModel, 100)

func broadcast() {

	sdp := <-offerChan
	// Everything below is the Pion WebRTC API, thanks for using it ❤️.
	offer := webrtc.SessionDescription{
		SDP:  sdp,
		Type: webrtc.SDPTypeOffer,
	}

	// Since we are answering use PayloadTypes declared by offerer
	mediaEngine := webrtc.MediaEngine{}
	err := mediaEngine.PopulateFromSDP(offer)
	if err != nil {
		panic(err)
	}

	mediaCodecs := mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeVideo)
	if len(mediaCodecs) == 0 {
		mediaCodecs = mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeAudio)
	}
	if len(mediaCodecs) == 0 {
		panic("no codec in offer")
	}

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.voipgate.com:3478", "stun:stun.ideasip.com"},
			},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
	}

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			log.Println(c.ToJSON())
		}
	})

	// Allow us to receive 1 video track
	// if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
	// 	log.Println("no video codec")
	// 	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
	// 		panic(err)
	// 	}
	// }

	//play back

	localTrack, err := peerConnection.NewTrack(mediaCodecs[0].PayloadType, rand.Uint32(), "video", "pion")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(localTrack)
	if err != nil {
		panic(err)
	}

	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}
	// localTrackChan := make(chan *webrtc.Track)
	// Set a handler for when a new remote track starts, this just distributes all our packets
	// to connected peers
	peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
		go func() {
			ticker := time.NewTicker(rtcpPLIInterval)
			for range ticker.C {
				if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}}); rtcpSendErr != nil {
					fmt.Println(rtcpSendErr)
				}
			}
		}()

		fmt.Printf("Track has started, of type %d: %s \n", remoteTrack.PayloadType(), remoteTrack.Codec().Name)
		// // Create a local track, all our SFU clients will be fed via this track
		// localTrack, newTrackErr := peerConnection.NewTrack(remoteTrack.PayloadType(), remoteTrack.SSRC(), "video", "pion")
		// if newTrackErr != nil {
		// 	panic(newTrackErr)
		// }
		// localTrackChan <- localTrack

		// rtpBuf := make([]byte, 1400)
		// for {
		// 	i, readErr := remoteTrack.Read(rtpBuf)
		// 	if readErr != nil {
		// 		panic(readErr)
		// 	}

		// 	// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
		// 	if _, err = localTrack.Write(rtpBuf[:i]); err != nil && err != io.ErrClosedPipe {
		// 		panic(err)
		// 	}
		// }
		for {
			// Read RTP packets being sent to Pion
			rtp, readErr := remoteTrack.ReadRTP()
			if readErr != nil {
				panic(readErr)
			}

			// Replace the SSRC with the SSRC of the outbound track.
			// The only change we are making replacing the SSRC, the RTP packets are unchanged otherwise
			rtp.SSRC = localTrack.SSRC()

			if writeErr := localTrack.WriteRTP(rtp); writeErr != nil {
				panic(writeErr)
			}
		}
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	// Set the remote SessionDescription

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	// Get the LocalDescription and take it to base64 so we can paste in browser
	answerChan <- answer.SDP

	//candidate
FOR:
	for {

		select {
		case cand := <-candChan:
			peerConnection.AddICECandidate(
				webrtc.ICECandidateInit{
					SDPMid:        &cand.SdpMid,
					SDPMLineIndex: &cand.SdpMlineindex,
					Candidate:     cand.Candidate,
				},
			)
		default:
			if peerConnection.ICEConnectionState() == webrtc.ICEConnectionStateConnected {
				break FOR
			}
		}
	}
	log.Println("waiting for output track")

	// localTrack := <-localTrackChan
	for {
		log.Println("start accept subscribler")
		sdp := <-offerChan
		recvOnlyOffer := webrtc.SessionDescription{
			SDP:  sdp,
			Type: webrtc.SDPTypeOffer,
		}

		// Create a new PeerConnection
		peerConnection, err := api.NewPeerConnection(config)
		if err != nil {
			panic(err)
		}

		peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			fmt.Printf("Connection State has changed %s \n", connectionState.String())
		})

		_, err = peerConnection.AddTrack(localTrack)
		if err != nil {
			panic(err)
		}

		// Set the remote SessionDescription
		err = peerConnection.SetRemoteDescription(recvOnlyOffer)
		if err != nil {
			panic(err)
		}

		// Create answer
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}

		// Sets the LocalDescription, and starts our UDP listeners
		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			panic(err)
		}

		// Get the LocalDescription and take it to base64 so we can paste in browser
		answerChan <- answer.SDP
	FOR1:
		for {
			select {
			case cand := <-candChan:
				peerConnection.AddICECandidate(
					webrtc.ICECandidateInit{
						SDPMid:        &cand.SdpMid,
						SDPMLineIndex: &cand.SdpMlineindex,
						Candidate:     cand.Candidate,
					},
				)
			default:
				if peerConnection.ICEConnectionState() == webrtc.ICEConnectionStateConnected {
					break FOR1
				}

			}

		}
	}
}
