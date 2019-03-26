package backend

import (
	"github.com/bradfitz/gomemcache/memcache"
)

// MemcacheClient blah
type MemcacheClient struct {
	client *memcache.Client
}

// User proto for serialization https://stackoverflow.com/questions/37618399/efficient-go-serialization-of-struct-to-disk

// NewMemcacheClient new client
func NewMemcacheClient(hostname string) CacheBackend {
	client := memcache.New(hostname)
	return &MemcacheClient{client: client}
}

// Get data from cache
func (m *MemcacheClient) Get(key string, result interface{}) error {
	cacheItem, err := m.client.Get(key)
	if err != nil {
		log.Error("failed fetching", key, err)
		return err
	}

	err = FromBytes(cacheItem.Value, result)
	if err != nil {
		log.Error("failed to deserialize memcached data for key", key, err)
		return err
	}

	return nil
}

// Set data in cache
func (m *MemcacheClient) Set(key string, data interface{}, ttl int) error {
	bytes, err := ToBytes(data)
	if err != nil {
		log.Error("failed to serialize memcached data for key", key, data, err)
		return err
	}

	item := memcache.Item{
		Key:        key,
		Value:      bytes,
		Expiration: int32(ttl),
	}
	return m.client.Set(&item)
}
