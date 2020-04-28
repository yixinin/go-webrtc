package room

import (
	"fmt"
	"go-webrtc/config"
	"go-webrtc/protocol"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-acme/lego/v3/log"
)

type HttpServer struct {
	sync.RWMutex
	config *config.Config
	rooms  map[int32]*Room
	room   *Room
	g      *gin.Engine
}

func NewHttpServer() *HttpServer {
	var c = &config.Config{
		Stun: []string{
			"stun:stun.voipgate.com:3478",
			"stun:stun.ideasip.com",
		},
	}
	var room = NewRoom(c, 1024, "hgfedcba87654321")
	var rooms = make(map[int32]*Room, 10)
	rooms[1024] = room

	var hs = &HttpServer{
		config: c,
		room:   room,
		rooms:  rooms,
		g:      gin.Default(),
	}
	hs.HandleHttp()
	return hs
}

func init() {
	NewHttpServer()
}

// func init() {
// 	DefaultConfig = &config.Config{
// 		Stun: []string{
// 			"stun:stun.voipgate.com:3478",
// 			"stun:stun.ideasip.com",
// 		},
// 	}
// 	DefaultRoom = NewRoom(DefaultConfig, 1024, "hgfedcba87654321")
// 	HandleHttp()
// }

func (hs *HttpServer) HandleHttp() {

	hs.g.StaticFS("/static", http.Dir("static"))
	hs.g.StaticFile("/index", "static/index.html")
	hs.g.StaticFile("/index.html", "static/index.html")

	hs.g.POST("/sendOffer", hs.SendOffer)
	hs.g.POST("/sendCandidate", hs.SendCandidate)
	// hs.g.GET("/pollCandidate", hs.PollCandidate)
	hs.g.GET("/createRoom", hs.CreateRoom)

	hs.g.GET("/getUsers", hs.GetUsers)

	hs.g.Run("0.0.0.0:8000")
}

type SendOfferModel struct {
	Offer   string `form:"offer" json:"offer"`
	FromUid int64  `form:"fromUid" json:"fromUid"`
	RoomId  int32  `form:"roomId" json:"roomId"`
	Uid     int64  `form:"uid" json:"uid"`
}

type SendCandidateModel struct {
	Uid       int64               `form:"uid" json:"uid"`
	Candidate *protocol.Candidate `form:"candidate" json:"candidate"`
	FromUid   int64               `form:"fromUid" json:"fromUid"`
	RoomId    int32               `form:"roomId" json:"roomId"`
}

type GetCandidateModel struct {
	Uid     int64 `form:"uid" json:"uid"`
	FromUid int64 `form:"fromUid" json:"fromUid"`
	RoomId  int32 `form:"roomId" json:"roomId"`
}

type CandiateModel struct {
	Candidate     string `json:"candidate" form:"candidate"`
	SdpMlineindex uint16 `json:"sdpMlineindex" form:"sdpMlineindex"`
	SdpMid        string `json:"sdpMid" form:"sdpMid"`
}

func (hs *HttpServer) SendOffer(c *gin.Context) {
	var p SendOfferModel
	err := c.ShouldBind(&p)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	var room, ok = hs.GetRoom(p.RoomId)
	if !ok && p.RoomId != 0 {
		c.String(400, "no such room")
		return
	}
	answer, err := room.NewPeerConnction(p.Uid, p.FromUid, p.Offer)
	if err != nil {
		log.Println(err)

		c.String(400, err.Error())
		return
	}
	c.String(200, answer)
}

func (hs *HttpServer) SendCandidate(c *gin.Context) {

	// var buf, err = ioutil.ReadAll(c.Request.Body)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	var p SendCandidateModel
	// err = json.Unmarshal(buf, &p)
	err := c.ShouldBindJSON(&p)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	if p.Candidate == nil {
		c.String(400, "candidate is null")
		return
	}
	var room, ok = hs.GetRoom(p.RoomId)
	if !ok && p.RoomId != 0 {
		c.String(400, "no such room")
		return
	}
	err = room.AddCandidate(p.Uid, p.FromUid, p.Candidate)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	c.String(200, "success")
}

type CreateRoomModel struct {
}

func (hs *HttpServer) CreateRoom(c *gin.Context) {
	var id = GetRoomID()
	hs.Lock()
	defer hs.Unlock()
	for {
		if _, ok := hs.rooms[id]; ok {
			id = GetRoomID()
		} else {
			break
		}
	}
	var room = NewRoom(hs.config, id, "hgfedcba87654321")
	hs.rooms[id] = room
	c.Redirect(302, fmt.Sprintf("/index?roomId=%d", id))
}

type PollCandidateModel struct {
	FromUid int64 `form:"fromUid" json:"fromUid"`
	RoomId  int32 `form:"roomId" json:"roomId"`
	Uid     int64 `form:"uid" json:"uid"`
}

func (hs *HttpServer) PollCandidate(c *gin.Context) {
	var p PollCandidateModel
	c.ShouldBind(&p)
	if p.FromUid == 0 && p.Uid == 0 {
		log.Println("pls fill uid and fromUid")
		c.String(400, "pls fill uid and fromUid")
		return
	}
	var room, ok = hs.GetRoom(p.RoomId)
	if !ok && p.RoomId != 0 {
		c.String(400, "no such room")
		return
	}
	var cand = room.GetCandidate(p.Uid, p.FromUid)
	if cand == nil {
		log.Println("no candidates")
		c.String(400, "no candidates")
		return
	}
	c.JSON(200, cand)
}

func (hs *HttpServer) GetRoom(id int32) (*Room, bool) {
	hs.RLock()
	defer hs.RUnlock()
	room, ok := hs.rooms[id]
	if !ok {
		room = hs.room
	}
	return room, ok
}

type GetUsersModel struct {
	RoomId int32
}

type UserModel struct {
	Uid      int64 `json:"uid"`
	HasVideo bool  `json:"hasVideo"`
	HasAudio bool  `json:"hasAudio"`
}

type GetUserAck struct {
	Users []*UserModel
}

func (hs *HttpServer) GetUsers(c *gin.Context) {
	var p GetUsersModel
	c.ShouldBind(&p)

	var room, ok = hs.GetRoom(p.RoomId)
	if !ok && p.RoomId != 0 {
		c.String(400, "no such room")
		return
	}

	var peers = room.GetPeers()

	var users = make([]*UserModel, 0, len(peers))
	for _, peer := range peers {
		users = append(users, &UserModel{
			HasAudio: peer.HasAudio(),
			HasVideo: peer.HasVideo(),
		})
	}
	var ack = &GetUserAck{
		Users: users,
	}
	c.JSON(200, ack)
}
