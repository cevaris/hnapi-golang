package main

import (
	"context"
	"fmt"

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
	cacheBackend backend.CacheClient
}

// NewCachedItemRepo cached backed item repository
func NewCachedItemRepo(itemBackend backend.ItemBackend, cacheBackend backend.CacheClient) ItemRepo {
	return &CachedItemRepo{
		itemBackend:  itemBackend,
		cacheBackend: cacheBackend,
	}
}

// Get cached items
func (c *CachedItemRepo) Get(ctx context.Context, itemIds []int) ([]model.Item, error) {
	itemChan, errChan := c.itemBackend.HydrateItem(ctx, itemIds)
	defer close(itemChan)
	defer close(errChan)

	items := make([]model.Item, 0)
	for range itemIds {
		select {
		case err, ok := <-errChan:
			fmt.Println("failed to hydrate item: ", err, ok)
			if err == context.Canceled {
				fmt.Println("hydrate item was cancelled: ", err, ok)
				break
			}
		case r, ok := <-itemChan:
			if !ok {
				fmt.Println("should not happen")
				continue
			}
			items = append(items, r)
		}
	}

	return items, nil
}
