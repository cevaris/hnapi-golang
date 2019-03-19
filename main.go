package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cevaris/hnapi/repo"
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

func items(w http.ResponseWriter, r *http.Request) {
	idsStr := r.URL.Query().Get("ids")
	if len(strings.Trim(idsStr, " ")) == 0 {
		http.Error(w, "missing ids parameter", 400)
		return
	}
	fmt.Println("found params", idsStr)

	strItemIds := strings.Split(idsStr, ",")
	itemIds := make([]int, len(strItemIds))
	for i, s := range strItemIds {
		num, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			fmt.Println("failed to parse " + s)
			continue
		}
		itemIds[i] = int(num)
	}
	fmt.Println("parsed ids", itemIds)

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	itemChan, errChan := repo.Hydrate(ctx, itemIds)

	items := make([]repo.Item, len(itemIds))
	for i := range itemIds {
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
			items[i] = r
		}
	}

	response := repo.Items{
		Data: items,
	}

	b, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		fmt.Println("failed to serialize json:", err)
	}
	w.Write(b)
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
