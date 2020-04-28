package room

import (
	"log"

	"github.com/pion/webrtc/v2"
)

type Peer struct {
	pub    *Publisher            //上传音视频流
	subs   map[int64]*Subscriber //下载音视频流
	uid    int64
	closed bool
}

func (p *Peer) Closed() bool {
	if p == nil {
		return true
	}
	if p.pub == nil || p.pub.Closed() {
		if len(p.subs) == 0 {
			return true
		}
		for _, v := range p.subs {
			if !v.closed {
				return false
			}
		}
		return true
	}
	return false
}

func NewPeer(uid int64) *Peer {
	return &Peer{
		uid:  uid,
		subs: make(map[int64]*Subscriber),
	}
}

func (p *Peer) AddPublisher(conn *webrtc.PeerConnection) {
	if p.pub != nil {
		p.pub.Close()
	}
	p.pub = NewPublisher(conn)
}
func (p *Peer) AddPubOutputTrack(track *webrtc.Track) {
	if p.pub != nil && p.pub.outputTracks != nil {
		p.pub.outputTracks = append(p.pub.outputTracks, track)
	} else {
		log.Println("peer conn closed")
	}

}

func (p *Peer) AddSubscriber(fromUid int64, conn *webrtc.PeerConnection, senders []*webrtc.RTPSender) {
	if sub, ok := p.subs[fromUid]; ok {
		if sub != nil {
			sub.Close()
		}
	}
	p.subs[fromUid] = NewSubscriber(fromUid, conn, senders)
}

func (p *Peer) Close() {
	if !p.pub.Closed() {
		p.pub.conn.Close()
	}
	for _, v := range p.subs {
		if !v.Closed() {
			v.conn.Close()
		}
	}
	p.pub = nil
	p.subs = make(map[int64]*Subscriber)
}

func (p *Peer) ClosePublisher() {
	if !p.pub.Closed() {
		p.pub.conn.Close()
	}
	p.pub = nil
}

func (p *Peer) CloseSubscrible() {
	for _, v := range p.subs {
		v.conn.Close()
	}
}

func (p *Peer) HasVideo() bool {
	if !p.HasPub() {
		return false
	}
	for _, v := range p.pub.outputTracks {
		if v.Codec().Type == webrtc.RTPCodecTypeVideo {
			return true
		}
	}
	return false
}
func (p *Peer) HasAudio() bool {
	if !p.HasPub() {
		return false
	}
	for _, v := range p.pub.outputTracks {
		if v.Codec().Type == webrtc.RTPCodecTypeAudio {
			return true
		}
	}
	return false
}

func (p *Peer) HasPub() bool {
	return !p.pub.Closed() && len(p.pub.outputTracks) > 0
}
