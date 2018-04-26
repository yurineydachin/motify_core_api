package searchengine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"godep.lzd.co/service/interfaces"
)

type SearchEngine struct {
	totalRequests uint64
	goodRequests  uint64
}

type ResultRow struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type gResponse struct {
	Results []ResultRow `json:"results"`
}

type status struct {
	Header []string   `json:"header"`
	Data   [][]string `json:"data"`
}

func (se *SearchEngine) Search(ctx context.Context, query string) ([]ResultRow, error) {
	se.totalRequests++

	res, err := http.Get("https://www.googleapis.com/customsearch/v1element?key=AIzaSyCVAXiUzRYsML1Pv6RwSG1gunmMikTzQqY&num=10&sig=432dd570d1a386253361f581254f9ca1&cx=014787010704180617188:xyc74jvhgve&q=" +
		url.QueryEscape(query))
	if err != nil {
		return nil, err
	}

	var response gResponse

	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	se.goodRequests++

	return response.Results, nil
}

func (se *SearchEngine) Caption() string {
	return "Search engine"
}

func (se *SearchEngine) Status() interfaces.Status {
	return interfaces.Status{
		Header: []string{"Total", "Good", "Good %"},
		Rows: []interfaces.StatusRow{
			{
				Data: []string{
					fmt.Sprint(se.totalRequests),
					fmt.Sprint(se.goodRequests),
					fmt.Sprintf("%02f", float64(se.goodRequests)/float64(se.totalRequests)*100),
				},
			},
		},
	}
}
