package aesgcm

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
)

func Decrypt(in []byte, hash string) ([]byte, error) {
	keyIv := make([]byte, hex.DecodedLen(len(hash)))
	_, err := hex.Decode(keyIv, []byte(hash))
	if err != nil {
		return nil, err
	}

	ivLen := len(keyIv) - 32

	block, err := aes.NewCipher(keyIv[ivLen:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, ivLen)
	if err != nil {
		return nil, err
	}

	decryptedContents, err := gcm.Open(make([]byte, 0, len(in)), keyIv[0:ivLen], in, nil)
	if err != nil {
		return nil, err
	}

	return decryptedContents, nil
}
