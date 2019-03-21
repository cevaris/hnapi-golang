package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Items is for serializin json
type Items struct {
	Items []Item `json:"items"`
}

// Item is either Story, Comment, or Poll
type Item struct {
	ID   int    `json:"id"`
	Type string `json:"type,omitempty"`
	By   string `json:"by,omitempty"`
	Time int    `json:"time,omitempty"`

	Deleted bool `json:"deleted,omitempty"`
	Dead    bool `json:"dead,omitempty"`

	Parent int `json:"parent,omitempty"`

	Poll  int   `json:"poll,omitempty"`
	Parts []int `json:"parts,omitempty"`

	Decendants int   `json:"decendants,omitempty"`
	Kids       []int `json:"kids,omitempty"`

	URL   string `json:"url,omitempty"`
	Score int    `json:"score,omitempty"`
	Title string `json:"title,omitempty"`
}

var myClient = &http.Client{Timeout: 10 * time.Second}

// ItemRepo hydrates Items
type ItemRepo interface {
	HydrateItem(ctx context.Context, itemIds []int) (chan Item, chan error)
}

// HydrateItem https://venilnoronha.io/designing-asynchronous-functions-with-go
func HydrateItem(ctx context.Context, itemIds []int) (chan Item, chan error) {
	itemChan := make(chan Item, len(itemIds))
	errChan := make(chan error, len(itemIds))

	for _, itemID := range itemIds {
		go asyncHydrate(ctx, itemID, itemChan, errChan)
	}
	return itemChan, errChan
}

func asyncHydrate(ctx context.Context, itemID int, itemChan chan<- Item, errChan chan<- error) {
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

	var item Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		fmt.Println("failed unmarshalling item", string(body), err)
		errChan <- err
	}

	itemChan <- item
}
