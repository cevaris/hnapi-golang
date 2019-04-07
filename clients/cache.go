package clients

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	"github.com/cevaris/timber"
)

// User proto for serialization https://stackoverflow.com/questions/37618399/efficient-go-serialization-of-struct-to-disk

var log = timber.NewGoogleLogger()

// CacheClient is the common cache interface
type CacheClient interface {
	Get(context.Context, string, interface{}) error
	MultiGet(context.Context, []string) ([][]byte, error)
	Set(context.Context, string, interface{}, time.Duration) error
}

// ToBytes niavely converts a value to []byte
func ToBytes(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	return buf.Bytes(), err
}

// FromBytes niavely converts bytes to some interface
func FromBytes(byteBuff []byte, result interface{}) error {
	buf := bytes.NewReader(byteBuff)
	enc := gob.NewDecoder(buf)
	return enc.Decode(result)
}
