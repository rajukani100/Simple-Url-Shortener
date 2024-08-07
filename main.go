package main

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

func main() {
	urlMap := make(map[string]urlData)
	mu := sync.Mutex{}
	router := setupRouter(urlMap, &mu)
	router.Run(":8080")
}

type urlData struct {
	Id          string `json:"id"`
	RedirectUrl string `json:"redirect_url"`
}

func setupRouter(urlMap map[string]urlData, mu *sync.Mutex) *gin.Engine {
	router := gin.Default()
	router.GET("/create", shortURL(urlMap, mu))
	router.GET("/:url", redirectUrl(urlMap))
	return router
}

func shortURL(urlMap map[string]urlData, mu *sync.Mutex) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.Query("url")
		if url == "" {
			c.String(http.StatusBadRequest, "Missing URL")
			return
		}

		hashValue := md5.Sum([]byte(url))
		hashString := hex.EncodeToString(hashValue[:])[:8]
		mu.Lock()
		defer mu.Unlock()
		urlMap[hashString] = urlData{
			Id:          hashString,
			RedirectUrl: url,
		}

		c.String(http.StatusOK, "Your Shortened URL: http://localhost:8080/%s", hashString)
	}
}

func redirectUrl(urlMap map[string]urlData) gin.HandlerFunc {
	return func(c *gin.Context) {
		hashString := c.Param("url")
		val, isExist := urlMap[hashString]
		if isExist {
			c.Redirect(http.StatusPermanentRedirect, val.RedirectUrl)
		} else {
			c.String(http.StatusNotFound, "URL not found")
		}
	}
}
