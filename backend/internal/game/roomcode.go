package game

import (
	"crypto/rand"
	"math/big"
)

// Room code configuration
const (
	// Characters excludes ambiguous characters (0/O, 1/I/L)
	roomCodeCharset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	roomCodeLength  = 6
)

// GenerateRoomCode creates a unique 6-character alphanumeric code
func GenerateRoomCode() string {
	code := make([]byte, roomCodeLength)
	charsetLen := big.NewInt(int64(len(roomCodeCharset)))

	for i := 0; i < roomCodeLength; i++ {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			// Fall back to a simple approach if crypto/rand fails
			// This should be extremely rare
			idx = big.NewInt(int64(i % len(roomCodeCharset)))
		}
		code[i] = roomCodeCharset[idx.Int64()]
	}

	return string(code)
}
