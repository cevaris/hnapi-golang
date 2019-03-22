package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cevaris/hnapi/model"
)

// ItemBackend hydrates Items
type ItemBackend interface {
	HydrateItem(ctx context.Context, itemIds []int) (chan model.Item, chan error)
}

// FireBaseItemBackend firebase backed http client
type FireBaseItemBackend struct {
	client *http.Client
}

// NewFireBaseItemBackend constructs a new item repo
func NewFireBaseItemBackend() *FireBaseItemBackend {
	return &FireBaseItemBackend{client: &http.Client{Timeout: 10 * time.Second}}
}

// HydrateItem https://venilnoronha.io/designing-asynchronous-functions-with-go
func (f *FireBaseItemBackend) HydrateItem(ctx context.Context, itemIds []int) (chan model.Item, chan error) {
	itemChan := make(chan model.Item, len(itemIds))
	errChan := make(chan error, len(itemIds))

	for _, itemID := range itemIds {
		go f.asyncHydrate(ctx, itemID, itemChan, errChan)
	}
	return itemChan, errChan
}

func (f *FireBaseItemBackend) asyncHydrate(ctx context.Context, itemID int, itemChan chan<- model.Item, errChan chan<- error) {
	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
		return // short circuit
	default:
	}

	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json?print=pretty", itemID)
	resp, err := f.client.Get(url)
	if err != nil {
		fmt.Println("failed making http request", url)
		errChan <- err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed reading http response", itemID)
		errChan <- err
	}

	fmt.Println("hydrated", itemID, string(body))

	var item model.Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		fmt.Println("failed unmarshalling item", string(body), err)
		errChan <- err
	}

	itemChan <- item
}
