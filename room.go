package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
)

const (
	rtcpPLIInterval = time.Second * 3
)

type Room struct {
	Id int32
	// api                  *webrtc.API
	peers map[int64]*PeerConnection

	// candidates map[int64]map[string]bool
	config webrtc.Configuration
}

type RoomCandidate struct {
	PubCandidate  []*CandiateModel
	RecvCandidate map[int64][]*CandiateModel
}

func NewRoom(c *Config) *Room {
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: c.Stun,
			},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
	}
	// m := webrtc.MediaEngine{}
	// var api = webrtc.NewAPI(webrtc.WithMediaEngine(m))
	return &Room{
		// api:                  api,
		// candidates: make(map[int64]map[string]bool),
		config: peerConnectionConfig,
		peers:  make(map[int64]*PeerConnection),
	}
}

func (r *Room) AddPeer(uid, fromUid int64, sdp string) (answerSdp string, err error) {
	offer := webrtc.SessionDescription{
		SDP:  sdp,
		Type: webrtc.SDPTypeOffer,
	}
	m := webrtc.MediaEngine{}
	var mediaCodecs []*webrtc.RTPCodec
	//发布
	// if fromUid == 0 {
	err = m.PopulateFromSDP(offer)
	if err != nil {
		return
	}
	var hasVideo = true
	var hasAudio = true
	mediaCodecs = m.GetCodecsByKind(webrtc.RTPCodecTypeVideo)
	if len(mediaCodecs) == 0 {
		hasVideo = false
		hasAudio = true
		mediaCodecs = m.GetCodecsByKind(webrtc.RTPCodecTypeAudio)
	}
	// }

	var api = webrtc.NewAPI(webrtc.WithMediaEngine(m))

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(r.config)
	if err != nil {
		return
	}

	var peer, ok = r.peers[uid]
	if !ok {
		peer = NewPeerConnection(uid)
		if len(mediaCodecs) == 0 {
			err = errors.New("no media published")
			return
		}
		outputTrack, err := peerConnection.NewTrack(mediaCodecs[0].PayloadType, rand.Uint32(), strconv.FormatInt(uid, 10), "pion")
		if err != nil {
			log.Println(err)
		}

		//TODO 测试 将视频流返回给发布者

		peer.Update(peerConnection, outputTrack)
		r.OnTrack(uid, peerConnection)

		if fromUid != 0 {
			//添加目标视频源
			targetPeer, ok := r.peers[fromUid]
			if ok {
				peer.AddTrack(fromUid, targetPeer.outputTrack)
			}
		} else {
			peerConnection.AddTrack(outputTrack)
			if hasVideo {
				// peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo)
			}
			if hasAudio {
				// peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio)
			}

		}

		r.peers[uid] = peer
	}

	// r.OnIceCandidate(uid, fromUid, peer)
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Println("connection state changed", state)
	})

	peerConnection.SetRemoteDescription(offer)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return
	}
	peerConnection.SetLocalDescription(answer)
	//TODO 将answer发送给客户端
	answerSdp = answer.SDP
	return
}

func (r *Room) AddCandidate(uid int64, m *CandiateModel) (err error) {
	if m == nil {
		err = errors.New("nil candidate")
		return
	}
	var candidate = webrtc.ICECandidateInit{
		Candidate:     m.Candidate,
		SDPMLineIndex: &m.SdpMlineindex,
		SDPMid:        &m.SdpMid,
	}

	var peer, ok = r.peers[uid]
	if !ok {
		return errors.New("you are not int room")
	}
	if peer.conn == nil {
		return errors.New("peer closed")
	}

	peer.conn.AddICECandidate(candidate)
	return
}

func (r *Room) OnTrack(uid int64, conn *webrtc.PeerConnection) {
	conn.OnTrack(func(inputTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := conn.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: inputTrack.SSRC()}})
				if errSend != nil {
					fmt.Println(errSend)
				}
			}
		}()

		var outputTrack = r.peers[uid].outputTrack
		fmt.Printf("Track has started, of type %d: %s \n", inputTrack.PayloadType(), inputTrack.Codec().Name)
		for {
			// Read RTP packets being sent to Pion
			rtp, readErr := inputTrack.ReadRTP()
			if readErr != nil {
				log.Println(readErr)
			}

			// Replace the SSRC with the SSRC of the outbound track.
			// The only change we are making replacing the SSRC, the RTP packets are unchanged otherwise
			rtp.SSRC = outputTrack.SSRC()

			if writeErr := outputTrack.WriteRTP(rtp); writeErr != nil {

			}
		}
	})
}

func (r *Room) CandidateSync() {
	for _, peer := range r.peers {
		if peer.conn != nil {
			if peer.conn.ICEConnectionState() != webrtc.ICEConnectionStateConnected {

			}
		}
	}
}

// func (r *Room) OnIceCandidate(uid, fromUid int64, peer *Peer) {
// 	if fromUid != 0 {
// 		if recv, ok := peer.recvs[fromUid]; ok {
// 			recv.conn.OnICECandidate(func(c *webrtc.ICECandidate) {
// 				if c == nil {
// 					return
// 				}
// 				var candiate = c.ToJSON()
// 				var m = CandiateModel{
// 					Candidate: candiate.Candidate,
// 				}
// 				if candiate.SDPMLineIndex != nil {
// 					m.SdpMlineindex = *candiate.SDPMLineIndex
// 				}
// 				if candiate.SDPMid != nil {
// 					m.SdpMid = *candiate.SDPMid
// 				}
// 				recv.candidate = append(recv.candidate, m)
// 				log.Println("candidate added, uid=", uid, "fromUid=", fromUid)
// 			})
// 		}
// 	} else {
// 		peer.pub.conn.OnICECandidate(func(c *webrtc.ICECandidate) {
// 			if c == nil {
// 				return
// 			}
// 			var candiate = c.ToJSON()
// 			var m = CandiateModel{
// 				Candidate: candiate.Candidate,
// 			}
// 			if candiate.SDPMLineIndex != nil {
// 				m.SdpMlineindex = *candiate.SDPMLineIndex
// 			}
// 			if candiate.SDPMid != nil {
// 				m.SdpMid = *candiate.SDPMid
// 			}
// 			peer.pub.candidate = append(peer.pub.candidate, m)

// 			log.Println("candidate added, uid=", uid)
// 		})
// 	}
// }

// func (r *Room) GetCandidate(uid, fromUid int64) (candiate []CandiateModel, err error) {

// 	var peer, ok = r.peers[uid]
// 	if !ok {
// 		err = errors.New("your are not int room")
// 		return
// 	}
// 	if peer.Closed() {
// 		err = errors.New("peer closed")
// 		return
// 	}
// 	if fromUid != 0 {
// 		if recv, ok := peer.recvs[fromUid]; ok && !recv.Closed() {
// 			candiate = recv.candidate
// 		}

// 	} else {
// 		if !peer.pub.Closed() {
// 			candiate = peer.pub.candidate
// 		}

// 	}
// 	return
// }
