package store

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// generateToken generates random four character strings of some
// selected characters, that are unlikely to be misread.
func generateToken() string {
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZ123456789"
	token := make([]byte, 4, 4)
	for i := 0; i < len(token); i++ {
		token[i] = chars[rand.Intn(len(chars))]
	}
	return string(token)
}
