package main

import "github.com/pion/webrtc/v2"

type Peer struct {
	pub    *Publisher            //上传音视频流
	subs   map[int64]*Subscriber //下载音视频流
	uid    int64
	closed bool
}

// type PeerConnection struct {
// 	uid         int64
// 	conn        *webrtc.PeerConnection
// 	outputTrack *webrtc.Track
// 	recvTracks  map[int64]*webrtc.Track
// }

// func NewPeerConnection(uid int64) *PeerConnection {
// 	return &PeerConnection{
// 		uid:        uid,
// 		recvTracks: make(map[int64]*webrtc.Track),
// 	}
// }

// func (p *PeerConnection) Update(conn *webrtc.PeerConnection, track *webrtc.Track) {
// 	p.conn = conn
// 	p.outputTrack = track
// }

// func (p *PeerConnection) AddTrack(fromUid int64, track *webrtc.Track) {
// 	if p.conn == nil {
// 		return
// 	}
// 	p.conn.AddTrack(track)
// 	p.recvTracks[fromUid] = track
// }

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
	p.pub = NewPublisher(conn)
}
func (p *Peer) AddTrack(track *webrtc.Track) {
	p.pub.outputTracks = append(p.pub.outputTracks, track)
}

func (p *Peer) AddSubscriber(fromUid int64, conn *webrtc.PeerConnection, senders []*webrtc.RTPSender) {
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
