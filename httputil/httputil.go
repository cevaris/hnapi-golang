package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var serverError = APIResponse{
	Status:  "error",
	Message: "server error",
}
var serverErrorJSONBytes, _ = marshal(serverError, true)
var serverErrorJSON = string(serverErrorJSONBytes)

// APIResponse wrapper for http responses
type APIResponse struct {
	Status  string      `json:"status,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func GetBool(r *http.Request, paramName string, defaultValue bool) (bool, error) {
	prettyJSONStr := r.URL.Query().Get("pretty")
	if len(prettyJSONStr) != 0 {
		value, err := strconv.ParseBool(prettyJSONStr)
		if err != nil {
			msg := fmt.Sprintf("failed to parse '%v' value of the param '%s', expected a boolean", prettyJSONStr, paramName)
			fmt.Println(msg, err.Error())
			return false, errors.New(msg)
		}
		fmt.Println("found pretty param", value)
		return value, nil
	}
	return defaultValue, nil
}

func GetSlice(r *http.Request, paramName string) ([]int, error) {
	value := r.URL.Query().Get(paramName)
	if len(value) > 0 {
		listOfStrings := strings.Split(value, ",")
		slice := make([]int, 0)
		for _, str := range listOfStrings {
			num, err := strconv.ParseInt(str, 10, 32)
			if err != nil {
				fmt.Println("failed to parse " + str)
				return nil, fmt.Errorf("failed to parse '%s', found in %v, expected a list", str, listOfStrings)
			}
			slice = append(slice, int(num))
		}
		fmt.Println("parsed slice", slice)
		return slice, nil
	}

	return nil, errors.New(paramName + " parameter not present or contains no value")
}

func SerializeErr(w http.ResponseWriter, err error) {
	response := APIResponse{Status: "error", Message: err.Error()}
	b, err := marshal(response, true)
	if err != nil {
		fmt.Println("failed to serialize json ", err, "for", response)
		http.Error(w, serverErrorJSON, 500)
		return
	}
	http.Error(w, string(b), 400)
}

func SerializeData(w http.ResponseWriter, data interface{}, isPrettyJSON bool) {
	response := APIResponse{Status: "ok", Data: data}
	b, err := marshal(response, isPrettyJSON)
	if err != nil {
		fmt.Println("failed to serialize json ", err, "for", response)
		http.Error(w, serverErrorJSON, 500)
		return
	}
	w.Write(b)
}

func marshal(data interface{}, prettyJSON bool) ([]byte, error) {
	if prettyJSON {
		return json.MarshalIndent(data, "", "    ")
	}
	return json.Marshal(data)
}
