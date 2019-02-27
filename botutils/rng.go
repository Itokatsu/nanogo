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

// return index
func RandWeights(weights []int) int {
	length := len(weights)
	addedWeights := make([]int, length)
	for idx, w := range weights {
		if idx == 0 {
			addedWeights[idx] = w
		} else {
			addedWeights[idx] = addedWeights[idx-1] + w
		}
	}

	r := RandInt(addedWeights[length-1]) //total weights
	for i, aw := range addedWeights {
		if r < aw {
			return i
		}
	}
	return length - 1
}