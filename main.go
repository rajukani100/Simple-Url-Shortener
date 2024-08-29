package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	collection := client.Database("shortner").Collection("urlInfo")
	router := setupRouter(*collection)
	initTemplates()
	router.Run(":8080")
}

type urlInfo struct {
	Id          string `bson:"id"`
	RedirectUrl string `bson:"redirect_url"`
}

var shortenTemp *template.Template

func initTemplates() {
	var error error
	shortenTemp, error = template.ParseFiles("shorten.html")
	if error != nil {
		fmt.Println(error)
	}

}

func setupRouter(coll mongo.Collection) *gin.Engine {
	router := gin.Default()
	router.StaticFile("/", "index.html")
	router.StaticFile("/short", "shorten.html")
	router.POST("/shorten", shortURL(coll))
	router.GET("/:url", redirectUrl(coll))
	return router
}

func shortURL(coll mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.PostForm("url")
		if url == "" {
			c.String(http.StatusBadRequest, "Missing URL")
			return
		}

		hashValue := md5.Sum([]byte(url))
		hashString := hex.EncodeToString(hashValue[:])[:8]

		var existingUrlInfo urlInfo
		filter := bson.D{{Key: "id", Value: hashString}}
		err := coll.FindOne(context.TODO(), filter).Decode(&existingUrlInfo)
		fmt.Println(err)
		if err == nil {
			shortenTemp.Execute(c.Writer, hashString)
			return
		} else if err != mongo.ErrNoDocuments {
			log.Println("Error checking for existing URL:", err)
			c.String(http.StatusInternalServerError, "Failed to shorten URL")
			return
		}

		urlData := urlInfo{Id: hashString, RedirectUrl: url}
		_, err = coll.InsertOne(context.TODO(), urlData)
		if err != nil {
			log.Println("Error inserting data:", err)
			c.String(http.StatusInternalServerError, "Failed to shorten URL")
			return
		}
		shortenTemp.Execute(c.Writer, hashString)

	}
}

func redirectUrl(coll mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {

		hashString := c.Param("url")
		filter := bson.D{{Key: "id", Value: bson.D{{Key: "$eq", Value: hashString}}}}

		var URL_INFO urlInfo
		err := coll.FindOne(context.TODO(), filter).Decode(&URL_INFO)
		if err != nil {
			fmt.Println(err)
		}

		c.Redirect(http.StatusPermanentRedirect, URL_INFO.RedirectUrl)
	}
}
