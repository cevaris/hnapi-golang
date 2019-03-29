package backend

import (
	"context"
	"fmt"
	"time"

	"github.com/cevaris/hnapi/clients"
	"github.com/cevaris/hnapi/model"
)

var cacheDurationTTL = time.Minute * time.Duration(1)

// var cacheDurationTTL = time.Second * time.Duration(3)

// ItemRepo hydrate me
type ItemRepo interface {
	Get(context.Context, []int) ([]model.Item, error)
}

// CachedItemRepo hydrates and caches Items
type CachedItemRepo struct {
	itemBackend  ItemBackend
	cacheBackend clients.CacheClient
}

// NewCachedItemRepo cached backed item repository
func NewCachedItemRepo(itemBackend ItemBackend, cacheBackend clients.CacheClient) ItemRepo {
	return &CachedItemRepo{
		itemBackend:  itemBackend,
		cacheBackend: cacheBackend,
	}
}

// Get cached items
func (c *CachedItemRepo) Get(ctx context.Context, itemIds []int) ([]model.Item, error) {
	log.Debug("%v", itemIds)
	resultItems := make([]model.Item, 0)

	needToHydrateItemIdsSet := make(map[int]bool, 0)
	keys := make([]string, len(itemIds))
	for _, ID := range itemIds {
		keys = append(keys, itemCacheKey(ID))
		needToHydrateItemIdsSet[ID] = true
	}
	log.Debug("cache keys to lookup %v", keys)

	cacheResultBytes, err := c.cacheBackend.MultiGet(keys)
	for _, itemBytes := range cacheResultBytes {
		var result model.Item
		err = clients.FromBytes(itemBytes, &result)
		if err != nil {
			log.Error("failed to deserialize %v", err)
		} else {
			log.Debug("cache hit %d", result.ID)
			resultItems = append(resultItems, result)
			delete(needToHydrateItemIdsSet, result.ID)
		}
	}

	needToHydrateItemIds := make([]int, 0)
	for ID := range needToHydrateItemIdsSet {
		needToHydrateItemIds = append(needToHydrateItemIds, ID)
	}

	itemChan, errChan := c.itemBackend.HydrateItem(ctx, needToHydrateItemIds)
	defer close(itemChan)
	defer close(errChan)

	log.Debug("items still needed to hydrate %v", needToHydrateItemIds)
	for range needToHydrateItemIds {
		select {
		case err, ok := <-errChan:
			if err == context.Canceled {
				log.Error("hydrate item was cancelled: %v %t", err, ok)
				continue
			}
			log.Error("failed to hydrate item %v %t", err, ok)

		case r, ok := <-itemChan:
			if !ok {
				log.Debug("should not happen %v", r)
				continue
			}

			key := itemCacheKey(r.ID)
			ttl := int(time.Now().UTC().Add(cacheDurationTTL).Unix())
			err := c.cacheBackend.Set(key, &r, ttl)
			if err != nil {
				log.Error("failed to write to cache %s %v", key, err)
			} else {
				log.Debug("wrote to cache %s", key)
			}

			resultItems = append(resultItems, r)
		}
	}

	return resultItems, nil
}

func itemCacheKey(id int) string {
	return fmt.Sprintf("item:%d", id)
}
