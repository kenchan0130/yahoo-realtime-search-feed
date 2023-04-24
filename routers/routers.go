package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kenchan0130/yahoo-realtime-search-feed/controllers"
	"github.com/kenchan0130/yahoo-realtime-search-feed/repositories"
)

func Init() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/health")
	})
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	yahooRealtimeSearchRepository := repositories.YahooRealtimeSearchRepository{
		HTTPClient: &http.Client{},
	}
	feedCtrl := controllers.FeedController{
		YahooRealtimeSearchRepository: yahooRealtimeSearchRepository,
	}
	r.HEAD("/feed", feedCtrl.Index)
	r.GET("/feed", feedCtrl.Index)

	return r
}
