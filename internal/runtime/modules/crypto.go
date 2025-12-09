package modules

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/dop251/goja"
	"github.com/levskiy0/m3m/pkg/schema"
)

type CryptoModule struct{}

func NewCryptoModule() *CryptoModule {
	return &CryptoModule{}
}

// Name returns the module name for JavaScript
func (c *CryptoModule) Name() string {
	return "$crypto"
}

// Register registers the module into the JavaScript VM
func (c *CryptoModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(c.Name(), map[string]interface{}{
		"md5":         c.MD5,
		"sha256":      c.SHA256,
		"randomBytes": c.RandomBytes,
	})
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
	if length <= 0 {
		return ""
	}
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GetSchema implements JSSchemaProvider
func (c *CryptoModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$crypto",
		Description: "Cryptographic utilities for hashing and random data",
		Methods: []schema.MethodSchema{
			{
				Name:        "md5",
				Description: "Generate MD5 hash of data",
				Params:      []schema.ParamSchema{{Name: "data", Type: "string", Description: "Data to hash"}},
				Returns:     &schema.ParamSchema{Type: "string"},
			},
			{
				Name:        "sha256",
				Description: "Generate SHA256 hash of data",
				Params:      []schema.ParamSchema{{Name: "data", Type: "string", Description: "Data to hash"}},
				Returns:     &schema.ParamSchema{Type: "string"},
			},
			{
				Name:        "randomBytes",
				Description: "Generate random bytes as hex string",
				Params:      []schema.ParamSchema{{Name: "length", Type: "number", Description: "Number of bytes to generate"}},
				Returns:     &schema.ParamSchema{Type: "string"},
			},
		},
	}
}

// GetCryptoSchema returns the crypto schema (static version)
func GetCryptoSchema() schema.ModuleSchema {
	return (&CryptoModule{}).GetSchema()
}
