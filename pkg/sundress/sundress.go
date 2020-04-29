package sundress

import "crypto/sha256"

func Hash(s string) [32]byte {
	return sha256.Sum256([]byte(s))
}
