package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
)

const (
	rtcpPLIInterval = time.Second * 3
)

type Room struct {
	Id  int32
	Key string
	// api                  *webrtc.API
	peers map[int64]*Peer

	// candidates map[int64]map[string]bool
	config webrtc.Configuration
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
		peers:  make(map[int64]*Peer),
	}
}

func (r *Room) AddPeer(uid, fromUid int64, sdp string) (answerSdp string, err error) {
	offer := webrtc.SessionDescription{
		SDP:  sdp,
		Type: webrtc.SDPTypeOffer,
	}
	var isPublisher = fromUid == 0
	m := webrtc.MediaEngine{}
	var mediaCodecs []*webrtc.RTPCodec
	//发布
	// if fromUid == 0 {
	err = m.PopulateFromSDP(offer)
	if err != nil {
		return
	}
	mediaCodecs = m.GetCodecsByKind(webrtc.RTPCodecTypeVideo)
	if len(mediaCodecs) == 0 {
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
		peer = NewPeer(uid)

		r.peers[uid] = peer
	}

	if isPublisher {
		peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo)
		peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio)

		peer.AddPublisher(peerConnection)

		r.OnTrack(uid, peerConnection)
	} else {
		//添加目标视频源

		targetPeer, ok := r.peers[fromUid]
		var senders = make([]*webrtc.RTPSender, 0, len(targetPeer.pub.outputTracks))
		if ok {
			for _, track := range targetPeer.pub.outputTracks {
				if track != nil {
					var sender *webrtc.RTPSender
					sender, err = peerConnection.AddTrack(track)
					if err != nil {
						log.Println(err)
						return
					}
					if sender != nil {
						senders = append(senders, sender)
					}
				}
			}
		}

		peer.AddSubscriber(fromUid, peerConnection, senders)
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
	if fromUid != 0 {
		if sub, ok := peer.subs[fromUid]; ok && !sub.Closed() {
			sub.conn.AddICECandidate(candidate)
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

		fmt.Printf("Track has started, of type %d: %s \n", inputTrack.PayloadType(), inputTrack.Codec().Name)
		//创建track
		var outputTrack, err = conn.NewTrack(inputTrack.PayloadType(), inputTrack.SSRC(), strconv.FormatInt(uid, 10), "output-"+inputTrack.Label())
		if err != nil {
			log.Println("create output track err", err)
			return
		}

		r.peers[uid].AddTrack(outputTrack)

		rtpBuf := make([]byte, 1400)
		for {
			i, readErr := inputTrack.Read(rtpBuf)
			if readErr != nil {
				log.Println(err)
				return
			}

			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if _, err = outputTrack.Write(rtpBuf[:i]); err != nil && err != io.ErrClosedPipe {
				log.Println(err)
				return
			}
		}
	})
}

func (r *Room) ControlSubscriber(uid, fromUid int64, videoOn, audioOn bool) {
	if peer, ok := r.peers[uid]; ok {
		if sub, ok := peer.subs[fromUid]; ok {
			if !videoOn && !audioOn {
				sub.Close()
				delete(peer.subs, fromUid)
			} else {
				for _, sender := range sub.senders {
					if !audioOn && sender.Track().Codec().Type == webrtc.RTPCodecTypeAudio {
						sub.conn.RemoveTrack(sender)
					}
					if !videoOn && sender.Track().Codec().Type == webrtc.RTPCodecTypeVideo {
						sub.conn.RemoveTrack(sender)
					}
				}
			}

		}
	}
}

func (r *Room) KickPublisher(uid int64) {
	if peer, ok := r.peers[uid]; ok {
		if peer.pub != nil {
			peer.pub.Close()
			peer.pub = nil
		}
		if len(peer.subs) == 0 {
			delete(r.peers, uid)
		}
	}
}
