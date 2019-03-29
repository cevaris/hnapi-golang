package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/cevaris/hnapi/clients"
	"github.com/cevaris/hnapi/model"
)

// ItemBackend hydrates Items
type ItemBackend interface {
	HydrateItem(ctx context.Context, itemIds []int) (chan model.Item, chan error)
}

// FireBaseItemBackend firebase backed http client
type FireBaseItemBackend struct {
	// client *pester.Client
	// client *http.Client
	client clients.HTTPClient
}

// NewFireBaseItemBackend constructs a new item repo
func NewFireBaseItemBackend(httpClient clients.HTTPClient) ItemBackend {
	// client := pester.New()
	// client.Concurrency = 1
	// client.MaxRetries = 5
	// client.Backoff = pester.ExponentialBackoff
	// var client = urlfetch.Client(ctx) // &http.Client{Timeout: 10 * time.Second}
	return &FireBaseItemBackend{client: httpClient}
}

// MAX http requests
// TODO move this outside of the backend, into a global helper method and pass into NewFireBaseItemBackend
const MAX = 25

var sem = make(chan int, MAX)

func incr() {
	sem <- 1
}
func decr() {
	<-sem
}

// HydrateItem https://venilnoronha.io/designing-asynchronous-functions-with-go
// Perhaps https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383
func (f *FireBaseItemBackend) HydrateItem(ctx context.Context, itemIds []int) (chan model.Item, chan error) {
	itemChan := make(chan model.Item, len(itemIds))
	errChan := make(chan error, len(itemIds))

	for _, itemID := range itemIds {
		log.Debug("processing=%d goroutines=%d", len(sem), runtime.NumGoroutine())
		incr()
		go f.asyncHydrate(ctx, itemID, itemChan, errChan)
	}
	return itemChan, errChan
}

func (f *FireBaseItemBackend) asyncHydrate(ctx context.Context, itemID int, itemChan chan<- model.Item, errChan chan<- error) {
	defer decr()

	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
		return // short circuit
	default:
	}

	log.Debug("%d fetching item", itemID)
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json?print=pretty", itemID)
	resp, err := f.client.Get(url)
	if err != nil {
		log.Error("failed making http request", url)
		errChan <- err
		return
	}
	defer resp.Body.Close()
	log.Debug("%d fetched items", itemID)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("failed reading http response", itemID)
		errChan <- err
		return
	}

	log.Debug("%d hydrated", itemID)

	var item model.Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		log.Error("failed unmarshalling item", string(body), err)
		errChan <- err
		return
	}

	itemChan <- item
	log.Debug("%d completed", itemID)
}
