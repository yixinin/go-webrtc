package room

import (
	"math/rand"
	"sync/atomic"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var roomId int32 = 10023

func GetRoomID() int32 {
	newRoomId := atomic.AddInt32(&roomId, rand.Int31n(10))
	return newRoomId
}
