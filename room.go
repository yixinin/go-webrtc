package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
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
	peers map[int64]*Peer

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
		SDPSemantics: webrtc.SDPSemanticsPlanB,
	}
	// m := webrtc.MediaEngine{}
	// var api = webrtc.NewAPI(webrtc.WithMediaEngine(m))
	return &Room{
		// api:                  api,
		// candidates: make(map[int64]map[string]bool),
		config: peerConnectionConfig,
		peers:  make(map[int64]*Peer),
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
	if fromUid == 0 {
		err = m.PopulateFromSDP(offer)
		if err != nil {
			return
		}
		mediaCodecs = m.GetCodecsByKind(webrtc.RTPCodecTypeVideo)
		if len(mediaCodecs) == 0 {
			mediaCodecs = m.GetCodecsByKind(webrtc.RTPCodecTypeAudio)
		}
	}

	var api = webrtc.NewAPI(webrtc.WithMediaEngine(m))

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(r.config)
	if err != nil {
		return
	}

	var peer, ok = r.peers[uid]
	if !ok {
		peer = NewPeer(uid)
		if fromUid == 0 {
			if len(mediaCodecs) == 0 {
				err = errors.New("no media published")
				return
			}
			outputTrack, err := peerConnection.NewTrack(mediaCodecs[0].PayloadType, rand.Uint32(), mediaCodecs[0].Name, "pion")
			if err != nil {
				log.Println(err)
			}

			//TODO 测试 将视频流返回给发布者
			peerConnection.AddTrack(outputTrack)

			peer.AddPublisher(api, peerConnection, outputTrack)
			r.OnTrack(uid, peerConnection)
		} else {
			//添加目标视频源
			targetPeer, ok := r.peers[fromUid]
			if ok {
				peerConnection.AddTrack(targetPeer.pub.outputTrack)
			}

			peer.AddReceiver(fromUid, peerConnection)
		}

		r.peers[uid] = peer
	}

	// r.OnIceCandidate(uid, fromUid, peer)

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

func (r *Room) AddCandidate(uid, fromUid int64, m *CandiateModel) (err error) {
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
	if peer.Closed() {
		return errors.New("peer closed")
	}
	if fromUid != 0 {
		if recv, ok := peer.recvs[fromUid]; ok && !recv.Closed() {
			recv.conn.AddICECandidate(candidate)
		}

	} else {
		if !peer.pub.Closed() {
			peer.pub.conn.AddICECandidate(candidate)
		}
	}

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

		var outputTrack = r.peers[uid].pub.outputTrack
		fmt.Printf("Track has started, of type %d: %s \n", inputTrack.PayloadType(), inputTrack.Codec().Name)
		for {
			// Read RTP packets being sent to Pion
			rtp, readErr := inputTrack.ReadRTP()
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
}

func (r *Room) CandidateSync() {
	for _, peer := range r.peers {
		if !peer.Closed() && peer.pub != nil && !peer.pub.Closed() {
			if peer.pub.conn.ICEConnectionState() != webrtc.ICEConnectionStateConnected {

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

func (r *Room) ClosePeer(uid int64) {
	if p, ok := r.peers[uid]; ok {
		p.Close()
		delete(r.peers, uid)
	}
}

func (r *Room) ClosePublisher(uid int64) {
	if p, ok := r.peers[uid]; ok {
		p.ClosePublisher()
	}
}
func (r *Room) CloseReceiver(uid, fromUid int64) {
	if p, ok := r.peers[uid]; ok {
		p.CloseReceiver(fromUid)
	}
}
