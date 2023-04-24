package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/kenchan0130/yahoo-realtime-search-feed/models"
	"github.com/kenchan0130/yahoo-realtime-search-feed/repositories"
	"github.com/samber/lo"
	"log"
	"net/http"
	"strings"
)

type FeedController struct {
	YahooRealtimeSearchRepository repositories.YahooRealtimeSearchRepository
}

func (fc FeedController) Index(c *gin.Context) {
	searchQuery := strings.TrimSpace(c.Query("q"))

	entryList, res, err := fc.YahooRealtimeSearchRepository.GetTimelineEntry(searchQuery)
	if err != nil {
		log.Println(err)
		c.String(http.StatusInternalServerError, "Error. Please check server log.")
		return
	}

	requestURL := ""
	if res != nil {
		requestURL = res.Request.URL.String()
	}

	rss, err := fc.generateFeed(entryList, searchQuery, requestURL)
	if err != nil {
		log.Println(err)
		c.String(http.StatusInternalServerError, "Error. Please check server log.")
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, rss)
}

func (fc FeedController) generateFeed(entryList *[]models.TimelineEntry, query string, requestURL string) (string, error) {
	feed := &feeds.Feed{
		Title:       fmt.Sprintf("Realtime Search Feed with '%s'", query),
		Link:        &feeds.Link{Href: requestURL},
		Description: fmt.Sprintf("This feed is Yahoo! Realtime Search with '%s'", query),
		Image: &feeds.Image{
			Link:  requestURL,
			Url:   "https://abs.twimg.com/responsive-web/web/icon-default.3c3b2244.png", // From https://twitter.com/manifest.json
			Title: fmt.Sprintf("Realtime Search Feed with '%s'", query),
		},
	}

	feed.Items = lo.Map(*entryList, func(entry models.TimelineEntry, _ int) *feeds.Item {
		return &feeds.Item{
			Title:       *entry.DisplayText,
			Link:        &feeds.Link{Href: *entry.URL},
			Description: *entry.URL,
			Created:     entry.CreatedAt.Time,
			Id:          *entry.ID,
		}
	})

	return feed.ToRss()
}
