package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cevaris/hnapi/data"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

// ItemRepo hydrates Items
type ItemRepo interface {
	HydrateItem(ctx context.Context, itemIds []int) (chan data.Item, chan error)
}

// FireBaseItemRepo firebase backed http client
type FireBaseItemRepo struct {
	client *http.Client
}

// NewFireBaseItemRepo constructs a new item repo
func NewFireBaseItemRepo() *FireBaseItemRepo {
	return &FireBaseItemRepo{client: &http.Client{Timeout: 10 * time.Second}}
}

// HydrateItem https://venilnoronha.io/designing-asynchronous-functions-with-go
func (f *FireBaseItemRepo) HydrateItem(ctx context.Context, itemIds []int) (chan data.Item, chan error) {
	itemChan := make(chan data.Item, len(itemIds))
	errChan := make(chan error, len(itemIds))

	for _, itemID := range itemIds {
		go asyncHydrate(ctx, itemID, itemChan, errChan)
	}
	return itemChan, errChan
}

func asyncHydrate(ctx context.Context, itemID int, itemChan chan<- data.Item, errChan chan<- error) {
	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
		return // short circuit
	default:
	}

	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json?print=pretty", itemID)
	resp, err := myClient.Get(url)
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

	var item data.Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		fmt.Println("failed unmarshalling item", string(body), err)
		errChan <- err
	}

	itemChan <- item
}
