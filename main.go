package main

import (
	"context"
	"encoding/json"
	"errors"
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

// APIResponse wrapper for http responses
type APIResponse struct {
	Status  string        `json:"status,omitempty"`
	Message string        `json:"message,omitempty"`
	Data    []interface{} `json:"data,omitempty"`
}

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

func requiredHTTPParam(w http.ResponseWriter, r *http.Request, paramName string) string {
	strValue := r.URL.Query().Get(paramName)
	if len(strValue) == 0 {
		http.Error(w, fmt.Sprintf("missing %s parameter", paramName), 400)
	} else {
		fmt.Println(fmt.Sprintf("found %s=%v", paramName, strValue))
	}
	return strValue
}

func getSlice(r *http.Request, paramName string) ([]int, error) {
	value := r.URL.Query().Get(paramName)
	if len(value) > 0 {
		listOfStrings := strings.Split(value, ",")
		slice := make([]int, 0)
		for _, str := range listOfStrings {
			num, err := strconv.ParseInt(str, 10, 32)
			if err != nil {
				fmt.Println("failed to parse " + str)
				return nil, fmt.Errorf("failed to parse '%s', found in %v", str, listOfStrings)
			}
			slice = append(slice, int(num))
		}
		fmt.Println("parsed slice", slice)
		return slice, nil
	}

	return nil, errors.New(paramName + " parameter not present or contains no value")
}

var serverError = APIResponse{
	Status:  "error",
	Message: "server error",
}
var serverErrorJSONBytes, _ = marshal(serverError, true)
var serverErrorJSON = string(serverErrorJSONBytes)

func items(w http.ResponseWriter, r *http.Request) {
	itemIds, err := getSlice(r, "ids")
	if err != nil {
		response := APIResponse{Status: "error", Message: err.Error()}
		b, err := marshal(response, true)
		if err != nil {
			fmt.Println("failed to serialize json ", err, "for", response)
			http.Error(w, serverErrorJSON, 500)
			return
		}
		http.Error(w, string(b), 400)
		return
	}

	var prettyJSON = false
	prettyJSONStr := r.URL.Query().Get("pretty")
	if len(prettyJSONStr) != 0 {
		value, err := strconv.ParseBool(prettyJSONStr)
		if err != nil {
			response := APIResponse{Status: "error", Message: err.Error()}
			b, err := marshal(response, true)
			if err != nil {
				fmt.Println("failed to serialize json:", err)
			}
			http.Error(w, string(b), 400)
			return
		}
		fmt.Println("found pretty param", prettyJSON)
		prettyJSON = value
	}

	// strItemIds := strings.Split(idsStr, ",")
	// itemIds := make([]int, 0)
	// for _, s := range strItemIds {
	// 	num, err := strconv.ParseInt(s, 10, 32)
	// 	if err != nil {
	// 		fmt.Println("failed to parse " + s)
	// 		continue
	// 	}
	// 	itemIds = append(itemIds, int(num))
	// }
	// fmt.Println("parsed ids", itemIds)

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	itemChan, errChan := repo.Hydrate(ctx, itemIds)
	defer close(itemChan)
	defer close(errChan)

	items := make([]repo.Item, 0)
	for range itemIds {
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
			items = append(items, r)
		}
	}

	response := repo.Items{
		Data: items,
	}

	b, err := marshal(response, prettyJSON)
	if err != nil {
		fmt.Println("failed to serialize json:", err)
	}
	w.Write(b)
}

func marshal(data interface{}, prettyJSON bool) ([]byte, error) {
	if prettyJSON {
		return json.MarshalIndent(data, "", "    ")
	}
	return json.Marshal(data)
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
