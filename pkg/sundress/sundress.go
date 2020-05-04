package sundress

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func Hash(s string) [32]byte {
	return sha256.Sum256([]byte(s))
}

// Local bdc encryption using our generated datakey
func BdcEncrypt(s string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(createHash(dataKey)))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, []byte(s), nil), nil
}

// Local bdc decryption using our generated datakey
func BdcDecrypt(d []byte) (*string, error) {
	block, err := aes.NewCipher([]byte(createHash(dataKey)))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := d[:nonceSize], d[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	res := new(string)
	*res = string(plaintext)
	return res, nil
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}
