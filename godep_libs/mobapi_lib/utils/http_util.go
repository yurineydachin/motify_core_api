package utils

import (
	"io/ioutil"
	"net/http"
	"strings"
)

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
