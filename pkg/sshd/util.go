package sshd

import (
	"crypto/md5"
	"encoding/hex"

	"golang.org/x/crypto/ssh"
)

// fingerprint generates a colon-separated fingerprint string from a public key.
func fingerprint(key ssh.PublicKey) string {
	hash := md5.Sum(key.Marshal())
	buf := make([]byte, hex.EncodedLen(len(hash)))
	hex.Encode(buf, hash[:])
	// We need this in colon notation:
	fp := make([]byte, len(buf)+15)

	i, j := 0, 0
	for ; i < len(buf); i++ {
		if i > 0 && i%2 == 0 {
			fp[j] = ':'
			j++
		}
		fp[j] = buf[i]
		j++
	}
	return string(fp)
}
