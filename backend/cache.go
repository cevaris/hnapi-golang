package backend

import (
	"bytes"
	"encoding/gob"
)

// CacheBackend is the common cache interface
type CacheBackend interface {
	Get(string, interface{}) error
	Set(interface{}, int) error
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
