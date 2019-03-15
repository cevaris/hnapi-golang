package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func hello(w http.ResponseWriter, r *http.Request) {
	hydrateTopItems()
	io.WriteString(w, "Hellow World!")
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

func main() {
	port := os.Getenv("PORT")
	http.HandleFunc("/", hello)
	http.HandleFunc("/feed/top", topItems)
	http.ListenAndServe(":"+port, nil)
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func hydrateTopItems() ([]int, error) {

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
