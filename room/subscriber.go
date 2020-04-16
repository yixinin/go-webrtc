package room

import "github.com/pion/webrtc/v2"

type Subscriber struct {
	fromUid int64
	conn    *webrtc.PeerConnection
	senders []*webrtc.RTPSender

	subscribled bool
	closed      bool
}

func NewSubscriber(fromUid int64, conn *webrtc.PeerConnection, senders []*webrtc.RTPSender) *Subscriber {
	return &Subscriber{
		fromUid:     fromUid,
		conn:        conn,
		senders:     senders,
		subscribled: true,
	}
}

func (s *Subscriber) AddTrack(track *webrtc.Track) error {
	sender, err := s.conn.AddTrack(track)
	s.senders = append(s.senders, sender)
	return err
}
func (s *Subscriber) Closed() bool {
	if s == nil {
		return true
	}
	if s.conn == nil {
		return true
	}
	if s.closed {
		return true
	}
	if s.conn.ConnectionState() == webrtc.PeerConnectionStateConnecting ||
		s.conn.ConnectionState() == webrtc.PeerConnectionStateConnected {
		return false
	}
	return true
}

func (s *Subscriber) Close() {
	if s.closed {
		return
	}
	if s.conn.ConnectionState() == webrtc.PeerConnectionStateConnected ||
		s.conn.ConnectionState() == webrtc.PeerConnectionStateConnecting {
		s.conn.Close()
	}
	s.closed = true
}
