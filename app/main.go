package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/cevaris/hnapi/clients"
	"github.com/cevaris/hnapi/logging"

	"github.com/cevaris/hnapi/backend"
	"github.com/cevaris/hnapi/httputil"
	"github.com/cevaris/hnapi/model"
	"github.com/cevaris/httprouter"
	"google.golang.org/appengine"

	"net/http/pprof"
	_ "net/http/pprof"
)

// var log = logging.NewLogger("main")
var log = logging.NewGoogleLogger()

// var itemRepo backend.ItemRepo

func topItems(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := appengine.NewContext(r)
	httpClient := clients.NewGoogleHTTPClient(ctx)
	itemBackend := backend.NewFireBaseItemBackend(httpClient)
	cacheBackend := clients.NewGoogleMemcacheClient()
	itemRepo := backend.NewCachedItemRepo(itemBackend, cacheBackend)

	isPrettyJSON, err := httputil.GetBool(r, "pretty", false)
	if err != nil {
		httputil.SerializeErr(w, err)
		return
	}

	log.Debug(ctx, "topItems found pretty param", isPrettyJSON)

	itemIds, err := hydrateTopItems(ctx)
	if err != nil {
		httputil.SerializeErr(w, errors.New("failed to fetch top item ids"))
		return
	}

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

func hydrateComments(ctx context.Context, itemRepo backend.ItemRepo, commentIds []int, results *[]model.Item, conversation *model.Conversation) error {
	if len(commentIds) == 0 {
		return nil
	}

	items, err := itemRepo.Get(ctx, commentIds)
	if err != nil {
		return err
	}

	for _, item := range items {
		*results = append(*results, item)

		newConversation := model.NewConversation(item.ID)
		hydrateComments(ctx, itemRepo, item.Kids, results, newConversation)
		conversation.Kids = append(conversation.Kids, newConversation)
	}

	// sort conversaton by provided comments list
	conversation.Kids = sortConversationByP(conversation.Kids, commentIds)

	return nil
}

func item(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var ctx = appengine.NewContext(r)
	httpClient := clients.NewGoogleHTTPClient(ctx)
	itemBackend := backend.NewFireBaseItemBackend(httpClient)
	cacheBackend := clients.NewGoogleMemcacheClient()
	itemRepo := backend.NewCachedItemRepo(itemBackend, cacheBackend)

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
	log.Info(ctx, "just log text")
	log.Info(ctx, "found pretty param", isPrettyJSON)
	log.Info(ctx, "one", "two", "three")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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
	ctx = appengine.NewContext(r)
	err = hydrateComments(ctx, itemRepo, item.Kids, &comments, &conversation)
	if err != nil {
		log.Error(ctx, "failed hydrating comments, got", len(comments), "of", len(item.Kids))
	}

	response := model.Items{
		Items:        items,
		Conversation: conversation,
		Comments:     sortItemsByTime(comments),
	}

	httputil.SerializeData(w, response, isPrettyJSON)
}

func items(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := appengine.NewContext(r)
	httpClient := clients.NewGoogleHTTPClient(ctx)
	itemBackend := backend.NewFireBaseItemBackend(httpClient)
	cacheBackend := clients.NewGoogleMemcacheClient()
	itemRepo := backend.NewCachedItemRepo(itemBackend, cacheBackend)

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
	log.Debug(ctx, "found pretty param", isPrettyJSON)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
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

func hydrateTopItems(ctx context.Context) ([]int, error) {
	httpClient := clients.NewGoogleHTTPClient(ctx)

	resp, err := httpClient.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
	if err != nil {
		log.Error(ctx, "failed to hydrate items", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(ctx, "failed to read to bytes", err)
		return nil, err
	}

	itemIds := make([]int, 0)
	jsonErr := json.Unmarshal(body, &itemIds)
	if jsonErr != nil {
		log.Error(ctx, "failed to unmarshall top itemids", err)
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

func init() {
	// domain := getenv("DOMAIN", "0.0.0.0")
	// port := os.Getenv("PORT")

	router := httprouter.New()
	router.GET("/feed/top", topItems)
	router.GET("/items/:ID", item)
	router.GET("/items", items)
	router.GET("/debug/pprof/goroutine", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) { pprof.Index(w, r) })
	http.Handle("/", router)
	// http.ListenAndServe(domain+":"+port, router)
	appengine.Main()
}
