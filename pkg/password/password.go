package password

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math/rand/v2"
	"strings"

	"github.com/the-maldridge/wordlist"
)

type Mode uint8

const (
	// PasswordModeRandomHex is used to specify that we want a
	// password comprised of a string of upper case hex
	// characters.
	PasswordModeRandomHex Mode = iota + 1

	// PasswordModeXKCD is used to specify that we want a password
	// comprised of a string of random words joined by hyphens, a
	// la https://xkcd.com/936/
	PasswordModeXKCD
)

type Password struct {
	components []string
	pass       string
	mode       Mode
	size       int
}

func New(components []string, mode Mode, size int) *Password {
	p := Password{components: components, mode: mode, size: size}

	p.generate()
	return &p
}

func (p *Password) String() string { return p.pass }

func (p *Password) generate() {
	hash := sha256.Sum256([]byte(strings.Join(p.components, "")))

	// Convert the first 16 bytes of the hash into two uint64 seeds
	seed1 := binary.BigEndian.Uint64(hash[0:8])
	seed2 := binary.BigEndian.Uint64(hash[8:16])

	// Create a new seeded generator
	pcg := rand.NewPCG(seed1, seed2)
	r := rand.New(pcg)

	switch p.mode {
	case PasswordModeRandomHex:
		p.generateRandomHex(r)
	case PasswordModeXKCD:
		p.generateXKCD(r)
	}
}

func (p *Password) generateRandomHex(r *rand.Rand) {
	b := make([]byte, p.size)
	for i := range b {
		b[i] = byte(r.UintN(256))
	}
	p.pass = strings.ToUpper(hex.EncodeToString(b))
}

func (p *Password) generateXKCD(r *rand.Rand) {
	parts := make([]string, p.size)
	max := len(wordlist.Words)
	for i := range parts {
		parts[i] = wordlist.Words[r.IntN(max)]
	}
	p.pass = strings.Join(parts, "-")
}
