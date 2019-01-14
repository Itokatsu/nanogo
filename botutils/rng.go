package botutils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Generator() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// return random number between 0 and max-1
func RandInt(max int) int {
	return Generator().Intn(max)
}
