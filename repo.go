package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cevaris/hnapi/backend"
	"github.com/cevaris/hnapi/model"
)

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

var set = make(map[int]bool)

// Get cached items
func (c *CachedItemRepo) Get(ctx context.Context, itemIds []int) ([]model.Item, error) {
	resultItems := make([]model.Item, 0)
	// needToHydrateItemIds := itemIds
	needToHydrateItemIds := make([]int, 0)

	for _, ID := range itemIds {
		key := itemCacheKey(ID)
		var item model.Item
		err := c.cacheBackend.Get(key, &item)
		if err != nil {
			// fmt.Println("cache miss", key, err)
			needToHydrateItemIds = append(needToHydrateItemIds, ID)
		} else {
			// fmt.Println("cache hit", key)
			resultItems = append(resultItems, item)
		}
	}

	for _, ID := range itemIds {
		set[ID] = true
	}

	itemChan, errChan := c.itemBackend.HydrateItem(ctx, needToHydrateItemIds)
	defer close(itemChan)
	defer close(errChan)

	// fmt.Println("items still needed to hydrate", needToHydrateItemIds)
	for range needToHydrateItemIds {
		select {
		case err, ok := <-errChan:
			if err == context.Canceled {
				fmt.Println("hydrate item was cancelled: ", err, ok)
				return resultItems, err
			}
			fmt.Println("failed to hydrate item: ", err, ok)

		case r, ok := <-itemChan:
			if !ok {
				fmt.Println("should not happen")
				continue
			}

			key := itemCacheKey(r.ID)
			ttl := int(time.Now().UTC().Add(time.Minute * time.Duration(10)).Unix())
			err := c.cacheBackend.Set(key, &r, ttl)
			if err != nil {
				// fmt.Println("failed to write to cache", key, err.Error())
			} else {
				// fmt.Println("wrote to cache", key)
			}

			resultItems = append(resultItems, r)
			delete(set, r.ID)
			// fmt.Println("completed", r.ID, "left over", len(set), set)
		}
	}

	return resultItems, nil
}

func itemCacheKey(id int) string {
	return fmt.Sprintf("item:%d", id)
}
