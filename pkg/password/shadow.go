package password

import (
	"errors"
	"os"

	"github.com/go-crypt/crypt/algorithm/shacrypt"
	"github.com/google/renameio/v2"
	"github.com/the-maldridge/shadow"
)

func (p *Password) UpdateShadow(path, user string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	sMap, err := shadow.ParseShadowMap(f)
	if err != nil {
		return err
	}

	hasher, err := shacrypt.New(shacrypt.WithSHA512())
	if err != nil {
		return err
	}
	digest, err := hasher.Hash(p.pass)
	if err != nil {
		return err
	}

	entries := sMap.FilterUID(func(s string) bool { return s == user })
	if len(entries) != 1 {
		return errors.New("uid does not match/does not match exactly")
	}
	sMap.Del([]*shadow.ShadowEntry{entries[0]})
	entries[0].Password = digest.String()
	sMap.Add([]*shadow.ShadowEntry{entries[0]})

	shadowBytes := []byte(sMap.String())
	if err := renameio.WriteFile(path, shadowBytes, 0400); err != nil {
		return err
	}
	return nil
}
