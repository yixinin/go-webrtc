package main

import (
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
	Uid       int64  `form:"uid" json:"uid"`
	Candidate string `form:"candidate" json:"candidate"`
	FromUid   int64  `form:"fromUid" json:"fromUid"`
	RoomId    int32  `form:"roomId" json:"roomId"`
}

type GetCandidateModel struct {
	Uid     int64 `form:"uid" json:"uid"`
	FromUid int64 `form:"fromUid" json:"fromUid"`
	RoomId  int32 `form:"roomId" json:"roomId"`
}

func GetAnswer(c *gin.Context) {
	var p GetAnswerModel
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
	var p SendCandidateModel
	err := c.ShouldBind(&p)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	err = DefaultRoom.AddCandidate(p.Uid, p.FromUid, p.Candidate)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	c.String(200, "success")
}

func GetCandidate(c *gin.Context) {
	var p GetCandidateModel
	err := c.ShouldBind(&p)
	if err != nil {
		c.String(400, err.Error())
		log.Println(err)
		return
	}
	candidate, err := DefaultRoom.GetCandidate(p.Uid, p.FromUid)
	if err != nil {
		log.Println(err)
		c.String(400, err.Error())
		return
	}
	c.JSON(200, candidate)
}
