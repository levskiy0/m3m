package modules

import (
	"encoding/base64"
	"encoding/json"
	"net/url"

	"github.com/dop251/goja"
)

type EncodingModule struct{}

func NewEncodingModule() *EncodingModule {
	return &EncodingModule{}
}

// Name returns the module name for JavaScript
func (e *EncodingModule) Name() string {
	return "encoding"
}

// Register registers the module into the JavaScript VM
func (e *EncodingModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(e.Name(), map[string]interface{}{
		"base64Encode":  e.Base64Encode,
		"base64Decode":  e.Base64Decode,
		"jsonParse":     e.JSONParse,
		"jsonStringify": e.JSONStringify,
		"urlEncode":     e.URLEncode,
		"urlDecode":     e.URLDecode,
	})
}

func (e *EncodingModule) Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func (e *EncodingModule) Base64Decode(data string) string {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return ""
	}
	return string(decoded)
}

func (e *EncodingModule) JSONParse(data string) interface{} {
	var result interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil
	}
	return result
}

func (e *EncodingModule) JSONStringify(data interface{}) string {
	result, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(result)
}

func (e *EncodingModule) URLEncode(data string) string {
	return url.QueryEscape(data)
}

func (e *EncodingModule) URLDecode(data string) string {
	decoded, err := url.QueryUnescape(data)
	if err != nil {
		return ""
	}
	return decoded
}

// GetSchema implements JSSchemaProvider
func (e *EncodingModule) GetSchema() JSModuleSchema {
	return JSModuleSchema{
		Name:        "encoding",
		Description: "Data encoding and decoding utilities",
		Methods: []JSMethodSchema{
			{
				Name:        "base64Encode",
				Description: "Encode data as base64 string",
				Params:      []JSParamSchema{{Name: "data", Type: "string", Description: "Data to encode"}},
				Returns:     &JSParamSchema{Type: "string"},
			},
			{
				Name:        "base64Decode",
				Description: "Decode base64 string to data",
				Params:      []JSParamSchema{{Name: "data", Type: "string", Description: "Base64 string to decode"}},
				Returns:     &JSParamSchema{Type: "string"},
			},
			{
				Name:        "jsonParse",
				Description: "Parse JSON string to object",
				Params:      []JSParamSchema{{Name: "data", Type: "string", Description: "JSON string to parse"}},
				Returns:     &JSParamSchema{Type: "any"},
			},
			{
				Name:        "jsonStringify",
				Description: "Convert object to JSON string",
				Params:      []JSParamSchema{{Name: "data", Type: "any", Description: "Object to stringify"}},
				Returns:     &JSParamSchema{Type: "string"},
			},
			{
				Name:        "urlEncode",
				Description: "URL-encode a string",
				Params:      []JSParamSchema{{Name: "data", Type: "string", Description: "String to encode"}},
				Returns:     &JSParamSchema{Type: "string"},
			},
			{
				Name:        "urlDecode",
				Description: "URL-decode a string",
				Params:      []JSParamSchema{{Name: "data", Type: "string", Description: "String to decode"}},
				Returns:     &JSParamSchema{Type: "string"},
			},
		},
	}
}

// GetEncodingSchema returns the encoding schema (static version)
func GetEncodingSchema() JSModuleSchema {
	return (&EncodingModule{}).GetSchema()
}
