package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/go-acme/lego/v3/log"
)

type GetAnswerModel struct {
	Offer   string `form:"offer" json:"offer"`
	FromUid int64  `form:"fromUid" json:"fromUid"`
	RoomId  int32  `form:"roomId" json:"roomId"`
	Uid     int64  `form:"uid" json:"uid"`
}

type SendCandidateModel struct {
	Uid       int64          `form:"uid" json:"uid"`
	Candidate *CandiateModel `form:"candidate" json:"candidate"`
	FromUid   int64          `form:"fromUid" json:"fromUid"`
	RoomId    int32          `form:"roomId" json:"roomId"`
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

func GetAnswer(c *gin.Context) {
	var p GetAnswerModel
	err := c.ShouldBind(&p)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	// answer, err := DefaultRoom.AddPeer(p.Uid, p.FromUid, p.Offer)
	// if err != nil {
	// 	log.Println(err)

	// 	c.String(400, err.Error())
	// 	return
	// }
	offerChan <- p.Offer

	answer := <-answerChan
	c.String(200, answer)
}

func SendCandidate(c *gin.Context) {

	var buf, err = ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println(err)
		return
	}
	// log.Println(string(buf))

	var p SendCandidateModel
	err = json.Unmarshal(buf, &p)
	// err = c.ShouldBind(&p)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	// log.Println("send candidate", p.Candidate)
	// err = DefaultRoom.AddCandidate(p.Uid, p.Candidate)
	// if err != nil {
	// 	log.Println(err)
	// 	c.String(400, err.Error())
	// 	return
	// }
	if p.Candidate == nil {
		c.String(400, "fail")
		return
	}
	candChan <- p.Candidate
	c.String(200, "success")
}

// func GetCandidate(c *gin.Context) {
// 	var p GetCandidateModel
// 	err := c.ShouldBind(&p)
// 	if err != nil {
// 		c.String(400, err.Error())
// 		log.Println(err)
// 		return
// 	}

// 	candidate, err := DefaultRoom.GetCandidate(p.Uid, p.FromUid)
// 	if err != nil {
// 		log.Println(err)
// 		c.String(400, err.Error())
// 		return
// 	}
// 	c.JSON(200, candidate)
// }

type SendSdpModel struct {
	Uid int64    `json:"uid"`
	Sdp SdpModel `json:"sdp"`
}

type PollSdpModel struct {
	FromUid int64  `json:"fromUid"`
	SdpType string `json:"sdpType"`
}

type PollCandModel struct {
	FromUid int64 `json:"fromUid"`
}

type SdpModel struct {
	Sdp     string `json:"sdp"`
	SdpType string `json:"sdpType"`
}

func PollSdp(c *gin.Context) {
	var p PollSdpModel
	c.ShouldBind(&p)
	log.Printf("%+v", p)
	// sdp := DefaultChat.GetSdp(p.FromUid, p.SdpType)
	sdp := DefaultReflect.answerSdp
	c.String(200, sdp)
}

func PollCandidate(c *gin.Context) {
	var p PollCandModel
	c.ShouldBind(&p)
	log.Printf("%+v", p)
	// candidates := DefaultChat.GetCandidate(p.FromUid)
	candidates := DefaultReflect.candidates
	if len(candidates) > 0 {
		c.JSON(200, candidates)
		return
	}
	c.String(400, "")
}

func SendSdp(c *gin.Context) {
	var p SendSdpModel
	c.ShouldBind(&p)
	log.Printf("%+v", p)
	DefaultChat.AddSdp(p.Uid, p.Sdp)
	c.String(200, "")
}

func SendCand(c *gin.Context) {
	var p SendCandidateModel
	c.ShouldBind(&p)
	log.Printf("%+v", p)
	DefaultChat.AddCandidate(p.Uid, p.Candidate)
	c.String(200, "")
}

type ReflectModel struct {
	Sdp string `json:"sdp" form:"sdp"`
}

func ReflectF(c *gin.Context) {
	var p ReflectModel
	c.ShouldBind(&p)

	offerChan <- p.Sdp

	sdp := <-answerChan
	log.Println("http recv answer")
	c.String(200, sdp)
}

func ReflectCand(c *gin.Context) {
	var p SendCandidateModel
	c.ShouldBind(&p)
	DefaultReflect.candidates = append(DefaultReflect.candidates, &ReflectCandidate{
		candidate: p.Candidate,
	})
}
