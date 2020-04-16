package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/go-acme/lego/v3/log"
)

type SendOfferModel struct {
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

func SendOffer(c *gin.Context) {
	var p SendOfferModel
	err := c.ShouldBind(&p)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	answer, err := DefaultRoom.AddPeer(p.Uid, p.FromUid, p.Offer)
	if err != nil {
		log.Println(err)

		c.String(400, err.Error())
		return
	}
	c.String(200, answer)
}

func SendCandidate(c *gin.Context) {

	var buf, err = ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var p SendCandidateModel
	err = json.Unmarshal(buf, &p)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	if p.Candidate == nil {
		c.String(400, "fail")
		return
	}
	// log.Println("send candidate", p.Candidate)
	err = DefaultRoom.AddCandidate(p.Uid, p.FromUid, p.Candidate)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	c.String(200, "success")
}

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

var offerChan = make(chan string)
var answerChan = make(chan string)
var candChan = make(chan *CandiateModel, 100)
var pollCandChan = make(chan *CandiateModel, 100)

func SendOfferChan(c *gin.Context) {
	var p SendOfferModel
	err := c.ShouldBind(&p)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	offerChan <- p.Offer

	answer := <-answerChan
	c.String(200, answer)
}

func SendCandChan(c *gin.Context) {
	var buf, err = ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println(err)
		return
	}
	// log.Println(string(buf))

	var p SendCandidateModel
	err = json.Unmarshal(buf, &p)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	if p.Candidate == nil {
		c.String(400, "fail")
		return
	}
	candChan <- p.Candidate
	c.String(200, "success")
}

func PollCandChan(c *gin.Context) {
	select {
	case cand := <-pollCandChan:
		c.JSON(200, cand)
	default:
		c.String(200, "")
	}

}
