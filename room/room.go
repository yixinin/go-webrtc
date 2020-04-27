package room

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"go-webrtc/config"
	"go-webrtc/protocol"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
)

const (
	rtcpPLIInterval = time.Second * 3
)

type Room struct {
	Id    int32
	Key   string
	peers map[int64]*Peer

	candidates map[int64]*PeerCandidate

	// candidates map[int64]map[string]bool
	config webrtc.Configuration
}

func NewRoom(c *config.Config, id int32, key string) *Room {
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
		Id:         id,
		Key:        key,
		config:     peerConnectionConfig,
		peers:      make(map[int64]*Peer),
		candidates: make(map[int64]*PeerCandidate),
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
		var senders []*webrtc.RTPSender
		if ok {
			senders = make([]*webrtc.RTPSender, 0, len(targetPeer.pub.outputTracks))
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
					log.Println("add track", sender.Track().Codec().Type)
				}
			}
		} else {
			log.Println("target not in room", fromUid)
		}

		peer.AddSubscriber(fromUid, peerConnection, senders)
	}

	// r.OnIceCandidate(uid, fromUid, peer)
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Println("connection state changed", state)
		if state == webrtc.PeerConnectionStateDisconnected ||
			state == webrtc.PeerConnectionStateClosed ||
			state == webrtc.ICETransportStateFailed {
			//TODO删除当前连接
			if _, ok := r.peers[uid]; ok {
				if isPublisher {
					r.peers[uid].pub = nil
					for _, peer := range r.peers {
						if len(peer.subs) > 0 {
							if sub, ok := peer.subs[uid]; ok {
								if sub != nil {
									peer.subs[uid].Close()
								}
								delete(peer.subs, uid)
							}
						}
					}
				} else {
					if len(r.peers[uid].subs) > 0 {
						if _, ok := r.peers[uid].subs[fromUid]; ok {
							delete(r.peers[uid].subs, fromUid)
						}
					}
				}
			}
		}
	})

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			var cand = c.ToJSON()

			var m = &protocol.Candidate{
				Candidate: cand.Candidate,
			}
			if cand.SDPMid != nil {
				m.SdpMid = *cand.SDPMid
			}
			if cand.SDPMLineIndex != nil {
				m.SdpMlineindex = uint32(*cand.SDPMLineIndex)
			}
			if fromUid == 0 {
				r.candidates[uid].AddPub(m, true)
			} else {
				r.candidates[uid].AddSub(fromUid, m, true)
			}
		}
	})

	peerConnection.SetRemoteDescription(offer)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return
	}
	peerConnection.SetLocalDescription(answer)
	//TODO 将answer发送给客户端
	answerSdp = answer.SDP
	fmt.Println(answerSdp)
	go r.SyncPeerCandidate(uid, fromUid, peerConnection)
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

		log.Printf("Track has started, of type %d: %s \n", inputTrack.PayloadType(), inputTrack.Codec().Type)
		//创建track
		var outputTrack, err = conn.NewTrack(inputTrack.PayloadType(), inputTrack.SSRC(), strconv.FormatInt(uid, 10), "output-"+inputTrack.Label())
		if err != nil {
			log.Println("create output track err", err)
			return
		}

		r.peers[uid].AddPubOutputTrack(outputTrack)
		//添加到所有订阅了当前用户的连接
		for _, peer := range r.peers {
			if len(peer.subs) > 0 {
				if sub, ok := peer.subs[uid]; ok {
					if !sub.Closed() {
						err := sub.AddTrack(outputTrack)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}

		}

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

func (r *Room) AddCandidate(uid, fromUid int64, m *protocol.Candidate) (err error) {
	if m == nil {
		err = errors.New("nil candidate")
		return
	}

	if _, ok := r.candidates[uid]; !ok {
		r.candidates[uid] = NewPeerCandidate()
	}
	if fromUid != 0 {
		// log.Println("save peer candidate, uid=", uid, "fromUid=", fromUid)
		r.candidates[uid].AddSub(fromUid, m, false)
	} else {
		// log.Println("save peer candidate, uid=", uid)
		r.candidates[uid].AddPub(m, false)
	}

	var smi = uint16(m.SdpMlineindex)
	var candidate = webrtc.ICECandidateInit{
		Candidate:     m.Candidate,
		SDPMLineIndex: &smi,
		SDPMid:        &m.SdpMid,
	}

	var peer, ok = r.peers[uid]
	if !ok {
		return nil
	}
	if fromUid != 0 {
		if sub, ok := peer.subs[fromUid]; ok && !sub.Closed() {
			sub.conn.AddICECandidate(candidate)
			// log.Println("add peer candidate", candidate)
		}
	} else {
		if !peer.pub.Closed() {
			peer.pub.conn.AddICECandidate(candidate)
			// log.Println("add peer candidate", candidate)
		}
	}
	return
}

func (r *Room) SyncPeerCandidate(uid, fromUid int64, conn *webrtc.PeerConnection) {
	for {

		if conn.ConnectionState() == webrtc.PeerConnectionStateConnected ||
			conn.ConnectionState() == webrtc.PeerConnectionStateFailed {
			break
		}

		if c, ok := r.candidates[uid]; ok {
			if fromUid != 0 {
				if sub, ok := c.subs[fromUid]; ok && len(sub.peer) > 0 {
					for _, v := range sub.peer {
						conn.AddICECandidate(v)
						// log.Println("add peer candidate", v)
					}
				}
			} else {
				if c.pub != nil && len(c.pub.peer) > 0 {
					for _, v := range c.pub.peer {
						conn.AddICECandidate(v)
						// log.Println("add peer candidate", v)
					}
				}

			}
		}
	}

}

func (r *Room) GetCandidate(uid, fromUid int64) []*protocol.Candidate {
	if c, ok := r.candidates[uid]; ok {
		if fromUid == 0 {
			if c.pub != nil && len(c.pub.local) > 0 {
				var items = make([]*protocol.Candidate, 0, len(c.pub.local))
				for _, v := range c.pub.local {
					var item = &protocol.Candidate{
						Candidate: v.Candidate,
					}
					if v.SDPMLineIndex != nil {
						item.SdpMlineindex = uint32(*v.SDPMLineIndex)
					} else {
						item.SdpMlineindex = 0
					}
					if v.SDPMid != nil {
						item.SdpMid = *v.SDPMid
					} else {
						item.SdpMid = ""
					}
					items = append(items, item)
				}
				return items
			}
		} else {
			if sub, ok := c.subs[fromUid]; ok && sub != nil && len(sub.local) > 0 {
				var items = make([]*protocol.Candidate, 0, len(sub.local))
				for _, v := range sub.local {
					var item = &protocol.Candidate{
						Candidate: v.Candidate,
					}
					if v.SDPMLineIndex != nil {
						item.SdpMlineindex = uint32(*v.SDPMLineIndex)
					}
					if v.SDPMid != nil {
						item.SdpMid = *v.SDPMid
					}
					items = append(items, item)
				}
				return items
			}
		}
	}
	return nil
}

func (r *Room) Control(uid, fromUid int64, videoOn, audioOn bool) {
	if peer, ok := r.peers[uid]; ok {
		if fromUid == 0 {
			r.controlPublisher(uid, peer, videoOn, audioOn)
			return
		}
		r.controlSubscriber(uid, fromUid, peer, videoOn, audioOn)
	}
}

func (r *Room) controlPublisher(uid int64, peer *Peer, videoOn, audioOn bool) {
	if !videoOn && !audioOn {
		//断开连接
		for _, v := range r.peers {
			if sub, ok := v.subs[uid]; ok && sub.subscribled && !sub.Closed() {
				sub.Close()
				delete(v.subs, uid)
			}
		}
		peer.pub = nil
	} else if !videoOn {
		r.RemoveTrack(uid, webrtc.RTPCodecTypeVideo)
	} else if !audioOn {
		r.RemoveTrack(uid, webrtc.RTPCodecTypeAudio)
	}
}

func (r *Room) controlSubscriber(uid, fromUid int64, peer *Peer, videoOn, audioOn bool) {
	if sub, ok := peer.subs[fromUid]; ok {
		//音视频都关闭 直接断开连接
		if !videoOn && !audioOn {
			sub.Close()
			sub.subscribled = false
			// delete(peer.subs, fromUid)
		} else if len(sub.senders) > 0 {

			var removes = make([]int, 0, len(sub.senders))
			for i, sender := range sub.senders {
				if sender == nil {
					continue
				}
				if !audioOn && sender.Track().Codec().Type == webrtc.RTPCodecTypeAudio {
					sub.conn.RemoveTrack(sender)
					removes = append(removes, i)
				}
				if !videoOn && sender.Track().Codec().Type == webrtc.RTPCodecTypeVideo {
					sub.conn.RemoveTrack(sender)
					removes = append(removes, i)
				}
			}
			//移除sender
			for i := len(removes) - 1; i >= 0; i-- {
				var index = removes[i]
				if index >= 0 && index < len(sub.senders) {
					sub.senders = append(sub.senders[:index], sub.senders[:index+1]...)
				}
			}
		}

	}
}

func (r *Room) RemoveTrack(uid int64, codeCType webrtc.RTPCodecType) {
	//移除pub track
	var removeIndex int
	for i, track := range r.peers[uid].pub.outputTracks {
		if track.Codec().Type == codeCType {
			removeIndex = i
			break
		}
	}
	if removeIndex >= 0 && removeIndex < len(r.peers[uid].pub.outputTracks) {
		r.peers[uid].pub.outputTracks = append(r.peers[uid].pub.outputTracks[:removeIndex], r.peers[uid].pub.outputTracks[removeIndex+1])
	}
	for _, v := range r.peers {
		if sub, ok := v.subs[uid]; ok && sub.subscribled && !sub.Closed() {
			//移除track
			for _, sender := range sub.senders {
				if sender.Track().Codec().Type == codeCType {
					sub.conn.RemoveTrack(sender)
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

func (r *Room) Kick(uid int64) {
	if peer, ok := r.peers[uid]; ok {
		if peer.pub != nil {
			peer.pub.Close()
			peer.pub = nil
		}
		if len(peer.subs) > 0 {
			for _, sub := range peer.subs {
				if sub != nil {
					sub.Close()
				}
			}
		}
		delete(r.peers, uid)
	}
	if len(r.peers) <= 1 {

		for _, peer := range r.peers {
			if peer.pub != nil {
				peer.pub.Close()
				peer.pub = nil
			}
			if len(peer.subs) > 0 {
				for _, sub := range peer.subs {
					if sub != nil {
						sub.Close()
					}
				}
			}
		}
		r.peers = make(map[int64]*Peer)
	}
}

func (r *Room) Close() {
	for uid := range r.peers {
		r.Kick(uid)
	}
}

func (r *Room) Closed() bool {
	return len(r.peers) == 0
}
