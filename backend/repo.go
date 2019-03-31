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
	log.Debug(ctx, itemIds)
	log.Info(ctx, "number of items to lookup", len(itemIds))
	resultItems := make([]model.Item, 0)

	needToHydrateItemIdsSet := make(map[int]bool, 0)
	keys := make([]string, 0)
	for _, ID := range itemIds {
		keys = append(keys, itemCacheKey(ID))
		needToHydrateItemIdsSet[ID] = true
	}
	log.Debug(ctx, "cache keys to lookup", keys)
	log.Info(ctx, "cache keys to lookup", len(keys))

	cacheResultBytes, err := c.cacheBackend.MultiGet(ctx, keys)
	for _, itemBytes := range cacheResultBytes {
		var result model.Item
		err = clients.FromBytes(itemBytes, &result)
		if err != nil {
			log.Error(ctx, "failed to deserialize", err)
		} else {
			log.Info(ctx, "cache hit", result.ID)
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

	log.Debug(ctx, "items still needed to hydrate", needToHydrateItemIds)
	log.Info(ctx, "items still needed to hydrate", len(needToHydrateItemIds))
	for range needToHydrateItemIds {
		select {
		case err, ok := <-errChan:
			if err == context.Canceled {
				log.Error(ctx, "hydrate item was cancelled", err, ok)
				continue
			}
			log.Error(ctx, "failed to hydrate item", err, ok)

		case r, ok := <-itemChan:
			if !ok {
				log.Debug(ctx, "should not happen", r)
				continue
			}

			key := itemCacheKey(r.ID)
			err := c.cacheBackend.Set(ctx, key, &r, cacheDurationTTL)
			if err != nil {
				log.Error(ctx, "failed to write to cache", key, err)
			} else {
				log.Debug(ctx, "wrote to cache", key)
			}

			resultItems = append(resultItems, r)
		}
	}

	return resultItems, nil
}

func itemCacheKey(id int) string {
	return fmt.Sprintf("item:%d", id)
}
