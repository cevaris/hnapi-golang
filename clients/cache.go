package clients

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	bmemcache "github.com/bradfitz/gomemcache/memcache"
	"github.com/cevaris/hnapi/logging"
	gmemcache "google.golang.org/appengine/memcache"
)

// User proto for serialization https://stackoverflow.com/questions/37618399/efficient-go-serialization-of-struct-to-disk

var log = logging.NewLogger("memcache")

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

// GoogleMemcacheClient blah
type googleMemcacheClient struct {
}

// NewGoogleMemcacheClient new client
// Delegates to client config to underlying google app engine memcache client
func NewGoogleMemcacheClient() CacheClient {
	return &googleMemcacheClient{}
}

// MultiGet data from cache
func (m *googleMemcacheClient) MultiGet(ctx context.Context, keys []string) ([][]byte, error) {
	cacheItemMap, err := gmemcache.GetMulti(ctx, keys)
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
func (m *googleMemcacheClient) Get(ctx context.Context, key string, result interface{}) error {
	cacheItem, err := gmemcache.Get(ctx, key)
	if err == bmemcache.ErrCacheMiss {
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
func (m *googleMemcacheClient) Set(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	bytes, err := ToBytes(data)
	if err != nil {
		log.Error("failed to serialize memcached data for key %s %v %v", key, data, err)
		return err
	}

	item := gmemcache.Item{
		Key:        key,
		Value:      bytes,
		Expiration: ttl,
	}
	return gmemcache.Set(ctx, &item)
}

// MemcacheClient blah
type bradfitzMemcacheClient struct {
	client *bmemcache.Client
}

// NewBradfitzMemcacheClient new client
func NewBradfitzMemcacheClient(hostname string) CacheClient {
	client := bmemcache.New(hostname)
	return &bradfitzMemcacheClient{client: client}
}

// MultiGet data from cache
func (m *bradfitzMemcacheClient) MultiGet(ctx context.Context, keys []string) ([][]byte, error) {
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
func (m *bradfitzMemcacheClient) Get(ctx context.Context, key string, result interface{}) error {
	cacheItem, err := m.client.Get(key)
	if err == bmemcache.ErrCacheMiss {
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
func (m *bradfitzMemcacheClient) Set(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	bytes, err := ToBytes(data)
	if err != nil {
		log.Error("failed to serialize memcached data for key %s %v %v", key, data, err)
		return err
	}

	item := bmemcache.Item{
		Key:        key,
		Value:      bytes,
		Expiration: int32(ttl),
	}
	return m.client.Set(&item)
}
