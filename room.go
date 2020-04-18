package main

import (
	"context"
	"go-webrtc/protocol"
	"go-webrtc/room"
	"log"
)

func (s *Server) CreateRoom(ctx context.Context, req *protocol.CreateRoomReq) (*protocol.CreateRoomAck, error) {
	var ack = &protocol.CreateRoomAck{}
	var id = room.GetRoomID()

	var room = room.NewRoom(s.config, id, req.Key)

	s.rooms[id] = room
	return ack, nil
}
func (s *Server) CloseRoom(ctx context.Context, req *protocol.CloseRoomReq) (*protocol.CloseRoomAck, error) {
	var ack = &protocol.CloseRoomAck{}
	if room, ok := s.rooms[req.Id]; ok {
		room.Close()
		delete(s.rooms, room.Id)
	}
	return ack, nil
}
func (s *Server) OpenPeer(ctx context.Context, req *protocol.OpenPeerReq) (*protocol.OpenPeerAck, error) {
	var ack = &protocol.OpenPeerAck{}
	room, ok := s.rooms[req.RoomId]
	if !ok {
		return ack, nil
	}
	var uid = room.GetUid(req.Key)
	if uid == 0 {
		return ack, nil
	}
	sdp, err := room.AddPeer(uid, req.FromUid, req.Sdp)
	if err != nil {
		log.Println(err)
	}
	ack.Sdp = sdp
	return ack, err
}
func (s *Server) AddCandidate(ctx context.Context, req *protocol.AddCandidateReq) (*protocol.AddCandidateAck, error) {
	var ack = &protocol.AddCandidateAck{}
	room, ok := s.rooms[req.RoomId]
	if !ok {
		ack.Code = 1
		return ack, nil
	}
	var uid = room.GetUid(req.Key)
	if uid == 0 {
		ack.Code = 2
		return ack, nil
	}
	err := room.AddCandidate(uid, req.FromUid, req.Candidate)
	if err != nil {
		ack.Code = 3
	}
	return ack, err
}
func (s *Server) Kick(ctx context.Context, req *protocol.KickReq) (*protocol.KickAck, error) {
	var ack = &protocol.KickAck{}
	var room, ok = s.rooms[req.RoomId]
	if !ok {
		ack.Code = 1
		return ack, nil
	}
	if req.Pub && !req.Sub {
		room.KickPublisher(req.Uid)
	} else {
		room.Kick(req.Uid)
	}
	//关闭房间
	if room.Closed() {
		delete(s.rooms, req.RoomId)
	}
	return ack, nil
}
func (s *Server) Control(ctx context.Context, req *protocol.ControlReq) (*protocol.ControlAck, error) {
	var ack = &protocol.ControlAck{}
	var room, ok = s.rooms[req.RoomId]
	if !ok {
		ack.Disconnected = false
		return ack, nil
	}
	room.Control(req.Uid, req.FromUid, req.VideoOn, req.AudioOn)
	return ack, nil
}
