package room

import (
	"go-webrtc/protocol"

	"github.com/pion/webrtc/v2"
)

type PeerCandidate struct {
	pub  *PeerCandidateItem
	subs map[int64]*PeerCandidateItem
}

type PeerCandidateItem struct {
	local []webrtc.ICECandidateInit
	peer  []webrtc.ICECandidateInit
}

func NewPeerCandidate() *PeerCandidate {
	return &PeerCandidate{
		pub: &PeerCandidateItem{
			local: make([]webrtc.ICECandidateInit, 0, 10),
			peer:  make([]webrtc.ICECandidateInit, 0, 10),
		},
		subs: make(map[int64]*PeerCandidateItem),
	}
}

func (p *PeerCandidate) AddPub(candidate *protocol.Candidate, local bool) {
	var smi = uint16(candidate.SdpMlineindex)
	var cand = webrtc.ICECandidateInit{
		Candidate:     candidate.Candidate,
		SDPMLineIndex: &smi,
		SDPMid:        &candidate.SdpMid,
	}
	if p.pub != nil {
		p.pub = &PeerCandidateItem{
			local: make([]webrtc.ICECandidateInit, 0, 10),
			peer:  make([]webrtc.ICECandidateInit, 0, 10),
		}
	}
	if local {
		p.pub.local = append(p.pub.local, cand)
	} else {
		p.pub.peer = append(p.pub.peer, cand)
	}

}

func (p *PeerCandidate) AddSub(fromUid int64, candidate *protocol.Candidate, local bool) {
	var smi = uint16(candidate.SdpMlineindex)
	var cand = webrtc.ICECandidateInit{
		Candidate:     candidate.Candidate,
		SDPMLineIndex: &smi,
		SDPMid:        &candidate.SdpMid,
	}
	if _, ok := p.subs[fromUid]; !ok {
		p.subs[fromUid] = &PeerCandidateItem{
			local: make([]webrtc.ICECandidateInit, 0, 10),
			peer:  make([]webrtc.ICECandidateInit, 0, 10),
		}
	}
	if local {
		p.subs[fromUid].local = append(p.subs[fromUid].local, cand)
	} else {
		p.subs[fromUid].peer = append(p.subs[fromUid].peer, cand)
	}

}
