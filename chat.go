package main

type Chat struct {
	peers map[int64]*ChatPeer
}

type ChatPeer struct {
	OfferSDp  string
	AnswerSdp string

	Candidate []*CandiateModel
}

func NewChat() *Chat {
	return &Chat{
		peers: make(map[int64]*ChatPeer),
	}
}

func (c *Chat) AddSdp(uid int64, sdp SdpModel) {
	peer, ok := c.peers[uid]
	if !ok {
		peer = &ChatPeer{
			Candidate: make([]*CandiateModel, 0),
		}
		c.peers[uid] = peer
	}
	switch sdp.SdpType {
	case "offer":
		peer.OfferSDp = sdp.Sdp
	case "answer":
		peer.AnswerSdp = sdp.Sdp
	}
}

func (c *Chat) AddCandidate(uid int64, candidate *CandiateModel) {
	peer, ok := c.peers[uid]
	if !ok {
		peer = &ChatPeer{
			Candidate: make([]*CandiateModel, 0),
		}
		c.peers[uid] = peer
	}
	peer.Candidate = append(peer.Candidate, candidate)
}

func (c *Chat) GetSdp(uid int64, sdpType string) string {
	if peer, ok := c.peers[uid]; ok {
		switch sdpType {
		case "offer":
			return peer.OfferSDp
		case "answer":
			return peer.AnswerSdp
		}
	}
	return ""
}

func (c *Chat) GetCandidate(uid int64) []*CandiateModel {
	if peer, ok := c.peers[uid]; ok {
		return peer.Candidate
	}
	return nil
}
