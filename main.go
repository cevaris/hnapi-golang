package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cevaris/hnapi/backend"
	"github.com/cevaris/hnapi/httputil"
	"github.com/cevaris/hnapi/model"
)

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World!")
}

func topItems(w http.ResponseWriter, r *http.Request) {
	itemIds, err := hydrateTopItems()
	if err != nil {
		io.WriteString(w, "[]")
	} else {
		response, err := json.Marshal(itemIds)
		if err != nil {
			io.WriteString(w, "[]")
		} else {
			io.WriteString(w, string(response))
		}
	}
}

var itemBackend = backend.NewFireBaseItemBackend()
var cacheBackend = backend.NewMemcacheClient("localhost:11211")
var itemRepo = NewCachedItemRepo(itemBackend, cacheBackend)

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
		Items: items,
	}

	httputil.SerializeData(w, response, isPrettyJSON)
}

func main() {
	domain := getenv("DOMAIN", "0.0.0.0")
	port := os.Getenv("PORT")
	http.HandleFunc("/", hello)
	http.HandleFunc("/feed/top", topItems)
	http.HandleFunc("/items", items)
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
