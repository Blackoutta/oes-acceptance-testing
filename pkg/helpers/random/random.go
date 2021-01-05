package random

import (
	"math/rand"
	"time"

	"github.com/rs/xid"
)

func RandString() string {
	guid := xid.NewWithTime(time.Now()).String()
	return guid[12:]
}

func RandInt(r int) int {
	s1 := rand.NewSource(time.Now().UnixNano() + rand.Int63())
	r1 := rand.New(s1)
	return r1.Intn(r)
}
