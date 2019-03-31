package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cevaris/hnapi/logging"
	"github.com/cevaris/httprouter"
)

var log = logging.NewLogger("httputil")

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

// GetBool parses http bool params
func GetBool(r *http.Request, paramName string, defaultValue bool) (bool, error) {
	boolValue := r.URL.Query().Get(paramName)
	if len(boolValue) != 0 {
		value, err := strconv.ParseBool(boolValue)
		if err != nil {
			msg := fmt.Sprintf("failed to parse '%v' value of the param '%s', expected a boolean", boolValue, paramName)
			log.Error("%s %s", msg, err.Error())
			return false, errors.New(msg)
		}
		return value, nil
	}
	return defaultValue, nil
}

// GetInt parses and returns a int
func GetInt(ps httprouter.Params, paramName string, defaultValue int) (int, error) {
	valueStr := ps.ByName(paramName)
	if len(valueStr) != 0 {
		value, err := strconv.ParseInt(valueStr, 10, 32)
		if err != nil {
			msg := fmt.Sprintf("failed to parse '%v' value of the param '%s', expected a boolean", valueStr, paramName)
			log.Error("%s %s", msg, err.Error())
			return defaultValue, errors.New(msg)
		}
		log.Debug("found pretty param %s", valueStr)
		return int(value), nil
	}
	return defaultValue, nil
}

// GetSlice parses http slices params
func GetSlice(r *http.Request, paramName string, defaultValue []int) ([]int, error) {
	value := r.URL.Query().Get(paramName)
	if len(value) > 0 {
		listOfStrings := strings.Split(value, ",")
		slice := make([]int, 0)
		for _, str := range listOfStrings {
			num, err := strconv.ParseInt(str, 10, 32)
			if err != nil {
				log.Error("failed to parse %s", str)
				return nil, fmt.Errorf("failed to parse '%s', found in %v, expected a list", str, listOfStrings)
			}
			slice = append(slice, int(num))
		}
		log.Debug("parsed slice", slice)
		return slice, nil
	}

	return defaultValue, nil
}

// SerializeErr writes exceptional JSON responses
func SerializeErr(w http.ResponseWriter, err error) {
	response := APIResponse{Status: "error", Message: err.Error()}
	b, err := marshal(response, true)
	if err != nil {
		log.Error("failed to serialize json %v for %v", err, response)
		http.Error(w, serverErrorJSON, 500)
		return
	}
	http.Error(w, string(b), 400)
}

// SerializeData writes data as JSON
func SerializeData(w http.ResponseWriter, data interface{}, isPrettyJSON bool) {
	response := APIResponse{Status: "ok", Data: data}
	b, err := marshal(response, isPrettyJSON)
	if err != nil {
		log.Error("failed to serialize json %v for %v", err, response)
		http.Error(w, serverErrorJSON, 500)
		return
	}
	w.Write(b)
	w.Header().Set("Content-Type", "application/json")
}

func marshal(data interface{}, prettyJSON bool) ([]byte, error) {
	if prettyJSON {
		return json.MarshalIndent(data, "", "    ")
	}
	return json.Marshal(data)
}
