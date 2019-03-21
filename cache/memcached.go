package cache

import "github.com/bradfitz/gomemcache/memcache"

// MemcacheClient blah
type MemcacheClient struct {
	client *memcache.Client
}

// User proto for serialization https://stackoverflow.com/questions/37618399/efficient-go-serialization-of-struct-to-disk

// NewMemcacheClient new client
func NewMemcacheClient(hostname string) *MemcacheClient {
	client := memcache.New(hostname)
	return &MemcacheClient{client: client}
}

// Get returns bytes for key
func (m *MemcacheClient) Get(key string) (interface{}, error) {
	item, err := m.client.Get(key)
	if err != nil {
		return nil, err
	}
	return item.Value, nil
}

// func (m *MemcacheClient) Set(interface{}, int) error {

// }
