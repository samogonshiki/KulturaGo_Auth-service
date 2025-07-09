package service

import (
	"crypto/rand"
	"crypto/subtle"
	"golang.org/x/crypto/argon2"
)

func salt() []byte { b := make([]byte, 16); _, _ = rand.Read(b); return b }

func hash(pwd string, s []byte) []byte {
	return append(s, argon2.IDKey([]byte(pwd), s, 1, 64*1024, 4, 32)...)
}
func verify(pwd string, h []byte) bool {
	s := h[:16]
	cmp := argon2.IDKey([]byte(pwd), s, 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(h[16:], cmp) == 1
}
