package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
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
func NewFireBaseItemBackend() ItemBackend {
	return &FireBaseItemBackend{client: &http.Client{Timeout: 5 * time.Second}}
}

// MAX http requests
const MAX = 40

var sem = make(chan int, MAX)

// HydrateItem https://venilnoronha.io/designing-asynchronous-functions-with-go
// Perhaps https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383
func (f *FireBaseItemBackend) HydrateItem(ctx context.Context, itemIds []int) (chan model.Item, chan error) {
	itemChan := make(chan model.Item, len(itemIds))
	errChan := make(chan error, len(itemIds))

	for _, itemID := range itemIds {
		fmt.Println(len(sem), runtime.NumGoroutine())
		sem <- 1
		go f.asyncHydrate(ctx, itemID, itemChan, errChan, sem)
	}
	return itemChan, errChan
}

func (f *FireBaseItemBackend) asyncHydrate(ctx context.Context, itemID int, itemChan chan<- model.Item, errChan chan<- error, sem <-chan int) {
	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
		return // short circuit
	default:
	}

	// fmt.Println(itemID, "fetching item")
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json?print=pretty", itemID)
	resp, err := f.client.Get(url)
	if err != nil {
		fmt.Println("failed making http request", url)
		errChan <- err
	}
	// fmt.Println(itemID, "fetched item")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed reading http response", itemID)
		errChan <- err
	}

	// fmt.Println("hydrated", itemID, string(body))
	// fmt.Println("hydrated", itemID)
	// fmt.Println(itemID, "hydrated")

	var item model.Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		fmt.Println("failed unmarshalling item", string(body), err)
		errChan <- err
	}

	itemChan <- item
	<-sem
	// fmt.Println(itemID, "completed item")
}
