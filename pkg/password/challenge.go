package password

import (
	"math/rand/v2"
	"strings"

	"github.com/the-maldridge/wordlist"
)

func ChallengeToken(size int) string {
	parts := make([]string, size)
	max := len(wordlist.Words)
	for i := range parts {
		parts[i] = wordlist.Words[rand.IntN(max)]
	}
	return strings.Join(parts, "-")
}
