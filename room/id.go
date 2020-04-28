package room

import (
	"math/rand"
	"time"
)

var R = rand.New(rand.NewSource(time.Now().Unix()))

func GetRoomID() int32 {
	newRoomId := 10023 + R.Int31()
	return newRoomId
}
