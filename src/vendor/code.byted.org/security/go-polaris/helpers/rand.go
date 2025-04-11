package helpers

import (
	"crypto/rand"
	"encoding/base64"
)

// RandBytes generates random bytes with fixed length.
func RandBytes(byteLen int) []byte {
	ret := make([]byte, byteLen)
	_, _ = rand.Read(ret)
	return ret
}

// RandBase64 geenrates random bytes in URL-safe base64 format.
func RandBase64(byteLen int) string {
	return base64.RawURLEncoding.EncodeToString(RandBytes(byteLen))
}
