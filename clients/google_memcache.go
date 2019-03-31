package clients

import (
	"context"
	"time"

	bmemcache "github.com/bradfitz/gomemcache/memcache"
	gmemcache "google.golang.org/appengine/memcache"
)

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
		log.Error(ctx, "failed fetching", keys, err)
		return nil, err
	}

	result := make([][]byte, len(cacheItemMap))
	var i = 0
	log.Info(ctx, "cache lookup found", len(cacheItemMap), "of", len(keys))
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
		log.Debug(ctx, "cache miss", key, err)
		return err
	} else if err != nil {
		log.Error(ctx, "failed fetching", key, err)
		return err
	}

	err = FromBytes(cacheItem.Value, result)
	if err != nil {
		log.Error(ctx, "failed to deserialize memcached data for key", key, err)
		return err
	}

	return nil
}

// Set data in cache
func (m *googleMemcacheClient) Set(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	bytes, err := ToBytes(data)
	if err != nil {
		log.Error(ctx, "failed to serialize memcached data for key", key, data, err)
		return err
	}

	item := gmemcache.Item{
		Key:        key,
		Value:      bytes,
		Expiration: ttl,
	}
	return gmemcache.Set(ctx, &item)
}
