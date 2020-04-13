package main

import "github.com/pion/webrtc/v2"

type Peer struct {
	pub   *Publisher          //上传音视频流
	recvs map[int64]*Receiver //下载音视频流
	uid   int64
}

func (p *Peer) Closed() bool {
	if p == nil {
		return true
	}
	if (p.pub == nil || p.pub.Closed()) && len(p.recvs) == 0 {
		return true
	}
	return false
}

type Receiver struct {
	uid       int64
	conn      *webrtc.PeerConnection
	candidate []string
}

func (r *Receiver) AddTrack(track *webrtc.Track) error {
	_, err := r.conn.AddTrack(track)
	return err
}

type Publisher struct {
	conn       *webrtc.PeerConnection
	api        *webrtc.API
	localTrack *webrtc.Track
	candidate  []string
}

func (p *Publisher) Closed() bool {
	if p == nil {
		return true
	}
	if p.conn == nil {
		return true
	}
	return p.conn.ConnectionState() == webrtc.PeerConnectionStateClosed
}
func (r *Receiver) Closed() bool {
	if r == nil {
		return true
	}
	if r.conn == nil {
		return true
	}
	return r.conn.ConnectionState() == webrtc.PeerConnectionStateClosed
}

func NewPeer(uid int64) *Peer {
	return &Peer{
		uid: uid,
	}
}

func (p *Peer) AddPublisher(api *webrtc.API, conn *webrtc.PeerConnection) {
	p.pub = &Publisher{
		conn:      conn,
		api:       api,
		candidate: make([]string, 0, 3),
	}
}

func (p *Peer) AddReceiver(fromUid int64, conn *webrtc.PeerConnection) {
	if _, ok := p.recvs[fromUid]; ok {
		p.recvs[fromUid].conn = conn
	}
	p.recvs[fromUid] = &Receiver{
		uid:       fromUid,
		conn:      conn,
		candidate: make([]string, 0, 3),
	}
}

func (p *Peer) Close() {
	if !p.pub.Closed() {
		p.pub.conn.Close()
	}
	for _, v := range p.recvs {
		if !v.Closed() {
			v.conn.Close()
		}
	}
	p.pub = nil
	p.recvs = make(map[int64]*Receiver)
}

func (p *Peer) ClosePublisher() {
	if !p.pub.Closed() {
		p.pub.conn.Close()
	}
	p.pub = nil
}

func (p *Peer) CloseReceiver(fromUid int64) {
	if v, ok := p.recvs[fromUid]; ok {
		if !v.Closed() {
			v.conn.Close()
		}
		p.recvs[fromUid] = nil
	}
}
