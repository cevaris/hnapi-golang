package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cevaris/hnapi/backend"
	"github.com/cevaris/hnapi/httputil"
	"github.com/cevaris/hnapi/model"
)

var itemRepo ItemRepo

func topItems(w http.ResponseWriter, r *http.Request) {
	isPrettyJSON, err := httputil.GetBool(r, "pretty", false)
	if err != nil {
		httputil.SerializeErr(w, err)
		return
	}

	itemIds, err := hydrateTopItems()
	if err != nil {
		httputil.SerializeErr(w, errors.New("failed to fetch top item ids"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	items, err := itemRepo.Get(ctx, itemIds)
	if err != nil {
		httputil.SerializeErr(w, err)
		return
	}

	response := model.Items{
		Items: sortItemsBy(items, itemIds),
	}

	httputil.SerializeData(w, response, isPrettyJSON)
}

func items(w http.ResponseWriter, r *http.Request) {
	itemIds, err := httputil.GetSlice(r, "ids", []int{})
	if err != nil {
		httputil.SerializeErr(w, err)
		return
	}

	if len(itemIds) == 0 {
		httputil.SerializeErr(w, errors.New("missing 'ids' parameter or values"))
		return
	}

	isPrettyJSON, err := httputil.GetBool(r, "pretty", false)
	if err != nil {
		httputil.SerializeErr(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	items, err := itemRepo.Get(ctx, itemIds)
	if err != nil {
		httputil.SerializeErr(w, err)
		return
	}

	response := model.Items{
		Items: sortItemsBy(items, itemIds),
	}

	httputil.SerializeData(w, response, isPrettyJSON)
}

func main() {
	domain := getenv("DOMAIN", "0.0.0.0")
	port := os.Getenv("PORT")
	cacheHostPort := getenv("CACHE_HOST", "localhost:11211")
	http.HandleFunc("/feed/top", topItems)
	http.HandleFunc("/items", items)

	itemBackend := backend.NewFireBaseItemBackend()
	cacheBackend := backend.NewMemcacheClient(cacheHostPort)
	itemRepo = NewCachedItemRepo(itemBackend, cacheBackend)

	http.ListenAndServe(domain+":"+port, nil)
}

func hydrateTopItems() ([]int, error) {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	resp, err := myClient.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println(string(body))

	itemIds := make([]int, 0)
	jsonErr := json.Unmarshal(body, &itemIds)
	if jsonErr != nil {
		log.Println(err)
		return nil, jsonErr
	}
	return itemIds, nil
}

func getenv(key, orElse string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return orElse
}

func sortItemsBy(source []model.Item, by []int) []model.Item {
	result := make([]model.Item, 0)
	for _, ID := range by {
		for _, v := range source {
			if v.ID == ID {
				result = append(result, v)
			}
		}
	}
	return result
}
