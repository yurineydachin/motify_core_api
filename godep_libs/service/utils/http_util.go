package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// DoAndUnmarshalGetRequest makes GET request and unmarshals response.
func DoAndUnmarshalGetRequest(client *http.Client, url string, data interface{}) error {
	body, err := DoGetRequest(client, url)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, data)
}

func DoGetRequest(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// ToURLValues converts map to url query.
func ToURLValues(params map[string]interface{}) url.Values {
	var values url.Values
	if len(params) > 0 {
		values = url.Values{}
		for k, v := range params {
			values.Set(k, fmt.Sprintf("%v", v))
		}
	}
	return values
}

func CreateRawURL(url, path string, values url.Values) string {
	rawURL := strings.TrimSuffix(url, "/") + "/" + strings.TrimPrefix(path, "/")
	if len(values) > 0 {
		rawURL += "?" + values.Encode()
	}
	return rawURL
}

func GetPath(req *http.Request) string {
	var path string
	if req != nil {
		path = req.URL.Path
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}
