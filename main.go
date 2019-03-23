package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/cevaris/hnapi/backend"
	"github.com/cevaris/hnapi/httputil"
	"github.com/cevaris/hnapi/model"
	"github.com/cevaris/httprouter"
)

var itemRepo ItemRepo

func topItems(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	isPrettyJSON, err := httputil.GetBool(r, "pretty", false)
	if err != nil {
		httputil.SerializeErr(w, err)
		return
	}
	fmt.Println("found pretty param", isPrettyJSON)

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

func hydrateComments(ctx context.Context, commentIds []int, results *[]model.Item, conversation *model.Conversation) error {
	items, err := itemRepo.Get(ctx, commentIds)
	if err != nil {
		return err
	}

	for _, item := range items {
		newConversation := model.NewConversation(item.ID)
		hydrateComments(ctx, item.Kids, results, newConversation)
		conversation.Kids = append(conversation.Kids, newConversation)

		if len(item.Kids) == 0 {
			continue
		}

		comments, err := itemRepo.Get(ctx, item.Kids)
		if err != nil {
			return err
		}

		*results = append(*results, comments...)
	}

	conversation.Kids = sortConversationByP(conversation.Kids, commentIds)

	return nil
}

func item(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	itemID, err := httputil.GetInt(ps, "ID", -1)
	if err != nil {
		httputil.SerializeErr(w, err)
		return
	}
	if itemID == -1 {
		httputil.SerializeErr(w, errors.New("missing parameter ':id'"))
		return
	}

	isPrettyJSON, err := httputil.GetBool(r, "pretty", false)
	if err != nil {
		httputil.SerializeErr(w, err)
		return
	}
	fmt.Println("found pretty param", isPrettyJSON)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	items, err := itemRepo.Get(ctx, []int{itemID})
	if err != nil {
		httputil.SerializeErr(w, err)
		return
	}

	var item model.Item
	if len(items) == 0 {
		httputil.SerializeErr(w, fmt.Errorf("failed to hydrate %d", itemID))
		return
	}
	item = items[0]

	comments := make([]model.Item, 0)
	conversation := model.Conversation{ID: itemID}
	err = hydrateComments(ctx, item.Kids, &comments, &conversation)

	response := model.Items{
		Items:        items,
		Conversation: conversation,
		Comments:     sortItemsByTime(comments),
	}

	httputil.SerializeData(w, response, isPrettyJSON)
}

func items(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	fmt.Println("found pretty param", isPrettyJSON)

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

	itemBackend := backend.NewFireBaseItemBackend()
	cacheBackend := backend.NewMemcacheClient(cacheHostPort)
	itemRepo = NewCachedItemRepo(itemBackend, cacheBackend)

	router := httprouter.New()
	router.GET("/feed/top", topItems)
	router.GET("/items/:ID", item)
	router.GET("/items", items)
	http.ListenAndServe(domain+":"+port, router)
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

func sortConversationByP(source []*model.Conversation, by []int) []*model.Conversation {
	result := make([]*model.Conversation, 0)
	for _, ID := range by {
		for _, v := range source {
			if v.ID == ID {
				result = append(result, v)
			}
		}
	}
	return result
}

func sortItemsByTime(source []model.Item) []model.Item {
	sort.Slice(source, func(i, j int) bool { return source[i].Time < source[j].Time })
	return source
}
