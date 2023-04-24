package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/kenchan0130/yahoo-realtime-search-feed/models"
)

type YahooRealtimeSearchRepository struct {
	HTTPClient *http.Client
}

func (t YahooRealtimeSearchRepository) GetTimelineEntry(query string) (*[]models.TimelineEntry, *http.Response, error) {
	if query == "" {
		return &[]models.TimelineEntry{}, nil, nil
	}

	u, err := url.Parse("https://search.yahoo.co.jp/realtime/search")
	if err != nil {
		return nil, nil, fmt.Errorf("url.Parse(): %v", err)
	}

	queryParameter := url.Values{
		"p":  {query},
		"ei": {"UTF-8"},
	}

	u.RawQuery = queryParameter.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, nil, fmt.Errorf("http.NewRequestWithContext(): %v", err)
	}

	res, err := t.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http.Client#Do(): %v", err)
	}
	defer req.Body.Close()

	if res.StatusCode != 200 {
		return nil, nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("goquery.NewDocumentFromReader(): %v", err)
	}

	dataStr := ""
	doc.Find("#__NEXT_DATA__").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		dataStr = selection.Text()
		return false
	})
	if dataStr == "" {
		return nil, nil, fmt.Errorf("'#__NEXT_DATA__' DOM is not found, please check %s response", u.String())
	}

	var data struct {
		Props *models.Props `json:"props,omitempty"`
	}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, nil, fmt.Errorf("json.Unmarshal(): %v", err)
	}

	if data.Props == nil {
		return nil, nil, fmt.Errorf("the data of Props is not found, please check %s response", u.String())
	}
	if data.Props.PageProps == nil {
		return nil, nil, fmt.Errorf("the data of Props.PageProps is not found, please check %s response", u.String())
	}
	if data.Props.PageProps.PageData == nil {
		return nil, nil, fmt.Errorf("the data of Props.PageProps.PageData is not found, please check %s response", u.String())
	}
	if data.Props.PageProps.PageData.Timeline == nil {
		return nil, nil, fmt.Errorf("the data of Props.PageProps.PageData.Timeline is not found, please check %s response", u.String())
	}

	return data.Props.PageProps.PageData.Timeline.Entry, res, nil
}
