package modules

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
)

type EncodingModule struct{}

func NewEncodingModule() *EncodingModule {
	return &EncodingModule{}
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
