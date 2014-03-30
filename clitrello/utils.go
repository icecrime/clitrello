package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func BuildURL(url string, params map[string]string) string {
	// Run a first pass to join each key-value URL parameter with an '=' char.
	paramList := make([]string, 0, len(params))
	for k, v := range params {
		paramList = append(paramList, strings.Join([]string{k, v}, "="))
	}

	// Build the final URL by concatenating the URL root, the request path and
	// the provided URL parameters (if any).
	resultUrl := url
	if len(paramList) != 0 {
		resultUrl += "?" + strings.Join(paramList, "&")
	}

	return resultUrl
}

func GetJSONContent(response *http.Response) []interface{} {
	var jsonData interface{}
	jsonDecoder := json.NewDecoder(response.Body)
	jsonDecoder.Decode(&jsonData)
	return jsonData.([]interface{})
}

func TrelloURL(config *Config, path string, params map[string]string) string {
	// We always have a minimum URL parameters to provide. God I miss Python's
	// or operator...
	if params == nil {
		params = make(map[string]string, 2)
	}

	// All requests to the Trello API must contains our application API key. We
	// also provide the user token when we have one.
	params["key"] = config.ApiKey
	if config.Token != "" {
		params["token"] = config.Token
	}

	endpoint := config.ApiRoot + path
	return BuildURL(endpoint, params)
}
