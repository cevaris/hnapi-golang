package clients

import (
	"bytes"
	"encoding/gob"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/cevaris/hnapi/logging"
)

var log = logging.NewLogger("memcache")

// CacheClient is the common cache interface
type CacheClient interface {
	Get(string, interface{}) error
	MultiGet([]string) ([][]byte, error)
	Set(string, interface{}, int) error
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

// MemcacheClient blah
type bradfitzMemcacheClient struct {
	client *memcache.Client
}

// User proto for serialization https://stackoverflow.com/questions/37618399/efficient-go-serialization-of-struct-to-disk
// NewMemcacheClient new client
func NewBradfitzMemcacheClient(hostname string) CacheClient {
	client := memcache.New(hostname)
	return &bradfitzMemcacheClient{client: client}
}

// MultiGet data from cache
func (m *bradfitzMemcacheClient) MultiGet(keys []string) ([][]byte, error) {
	cacheItemMap, err := m.client.GetMulti(keys)
	if err != nil {
		log.Error("failed fetching %s %v", keys, err)
		return nil, err
	}

	result := make([][]byte, len(cacheItemMap))
	var i = 0
	for _, cacheItem := range cacheItemMap {
		result[i] = cacheItem.Value
		i++
	}

	return result, nil
}

// Get data from cache
func (m *bradfitzMemcacheClient) Get(key string, result interface{}) error {
	cacheItem, err := m.client.Get(key)
	if err == memcache.ErrCacheMiss {
		log.Warning("cache miss %s %v", key, err)
		return err
	} else if err != nil {
		log.Error("failed fetching %s %v", key, err)
		return err
	}

	err = FromBytes(cacheItem.Value, result)
	if err != nil {
		log.Error("failed to deserialize memcached data for key %s %v", key, err)
		return err
	}

	return nil
}

// Set data in cache
func (m *bradfitzMemcacheClient) Set(key string, data interface{}, ttl int) error {
	bytes, err := ToBytes(data)
	if err != nil {
		log.Error("failed to serialize memcached data for key %s %v %v", key, data, err)
		return err
	}

	item := memcache.Item{
		Key:        key,
		Value:      bytes,
		Expiration: int32(ttl),
	}
	return m.client.Set(&item)
}
