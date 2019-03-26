package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cevaris/hnapi/backend"
	"github.com/cevaris/hnapi/model"
)

// var cacheDurationTTL = time.Minute * time.Duration(10)
var cacheDurationTTL = time.Second * time.Duration(3)

// ItemRepo hydrate me
type ItemRepo interface {
	Get(context.Context, []int) ([]model.Item, error)
}

// CachedItemRepo hydrates and caches Items
type CachedItemRepo struct {
	itemBackend  backend.ItemBackend
	cacheBackend backend.CacheBackend
}

// NewCachedItemRepo cached backed item repository
func NewCachedItemRepo(itemBackend backend.ItemBackend, cacheBackend backend.CacheBackend) ItemRepo {
	return &CachedItemRepo{
		itemBackend:  itemBackend,
		cacheBackend: cacheBackend,
	}
}

// Get cached items
func (c *CachedItemRepo) Get(ctx context.Context, itemIds []int) ([]model.Item, error) {
	log.Debug("%v", itemIds)
	resultItems := make([]model.Item, 0)
	needToHydrateItemIds := make([]int, 0)
	// needToHydrateItemIds := itemIds

	for _, ID := range itemIds {
		key := itemCacheKey(ID)
		var item model.Item
		err := c.cacheBackend.Get(key, &item)
		if err != nil {
			log.Error("cache miss %s %v", key, err)
			needToHydrateItemIds = append(needToHydrateItemIds, ID)
		} else {
			log.Debug("cache hit %s", key)
			resultItems = append(resultItems, item)
		}
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
