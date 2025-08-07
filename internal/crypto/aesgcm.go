package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"kulturago/auth-service/internal/custom_err"
)

func Encrypt(key, plain []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, custom_err.ErrKeySize
	}
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return append(nonce, gcm.Seal(nil, nonce, plain, nil)...), nil
}

func Decrypt(key, data []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, custom_err.ErrKeySize
	}
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	ns := gcm.NonceSize()
	if len(data) < ns {
		return nil, errors.New("cipher too short")
	}

	return gcm.Open(nil, data[:ns], data[ns:], nil)
}
