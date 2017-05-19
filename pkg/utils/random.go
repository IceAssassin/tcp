package utils

import (
	"math/rand"
	"time"
)

func getRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
func GetRandom(min, max int) int {
	r := getRand()
	return r.Intn(max-min) + min + 1
}

func GetRandomMax(max int) int {
	return GetRandom(0, max)
}
