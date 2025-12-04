package modules

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

type CryptoModule struct{}

func NewCryptoModule() *CryptoModule {
	return &CryptoModule{}
}

func (c *CryptoModule) MD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (c *CryptoModule) SHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (c *CryptoModule) RandomBytes(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
