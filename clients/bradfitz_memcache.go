package clients

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
		log.Error(ctx, "failed fetching", keys, err)
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
func (m *bradfitzMemcacheClient) Set(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	bytes, err := ToBytes(data)
	if err != nil {
		log.Error(ctx, "failed to serialize memcached data for key", key, data, err)
		return err
	}

	item := bmemcache.Item{
		Key:        key,
		Value:      bytes,
		Expiration: int32(ttl),
	}
	return m.client.Set(&item)
}
