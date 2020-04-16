package room

import "github.com/pion/webrtc/v2"

type Publisher struct {
	conn         *webrtc.PeerConnection
	outputTracks []*webrtc.Track

	closed bool
}

func NewPublisher(conn *webrtc.PeerConnection) *Publisher {
	return &Publisher{
		conn:         conn,
		outputTracks: make([]*webrtc.Track, 0, 2),
	}
}

func (p *Publisher) Close() {
	if p.closed {
		return
	}
	if p.conn.ConnectionState() == webrtc.PeerConnectionStateConnected ||
		p.conn.ConnectionState() == webrtc.PeerConnectionStateConnecting {
		p.conn.Close()
	}
	p.closed = true
}

func (p *Publisher) Closed() bool {
	if p == nil {
		return true
	}
	if p.conn == nil {
		return true
	}
	if p.closed {
		return true
	}
	if p.conn.ConnectionState() == webrtc.PeerConnectionStateConnecting ||
		p.conn.ConnectionState() == webrtc.PeerConnectionStateConnected {
		return false
	}
	return true
}
