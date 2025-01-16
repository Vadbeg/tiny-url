package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"tiny-url/storage"
	"tiny-url/utils"
)

var database *storage.SQLiteDatabase

type URLPayload struct {
	URL string `json:"url" binding:"required,url"`
}

func responseWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(
		statusCode,
		gin.H{
			"error": message,
		},
	)
}

func createShortLink(c *gin.Context) {
	var payload URLPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		responseWithError(c, http.StatusBadRequest, "Failed to parse URL")
		return
	}

	url := payload.URL
	fmt.Println(url)

	// Check if given link already exists in DB
	// If exists, return value retrieved from DB
	urlPostfix, err := database.GetShortURL(url)
	fmt.Printf("URL postfix retrieved from DB %v, with err %v\n", urlPostfix, err)

	if err == nil {
		c.JSON(
			http.StatusOK,
			gin.H{
				"message": "URL recieved succesfully",
				"url": url,
				"short_postfix": urlPostfix,
			},
		)
		return
	}

	urlPostfix = utils.ShortenUrl(url, 8)
	fmt.Printf("\nShorter url postfix %s for full url %s", urlPostfix, url)

	err = database.InsertURL(urlPostfix, url)
	if err != nil {
		fmt.Println(err)
		responseWithError(c, http.StatusNotFound, "Failed to load URL from DB")
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "URL recieved succesfully",
			"url": url,
			"short_postfix": urlPostfix,
		},
	)
}

func redirectByShortLink(c *gin.Context) {
	shortHash := c.Param("shortHash")

	fullUrl, err := database.GetFullURL(shortHash)

	if err != nil {
		responseWithError(c, http.StatusBadRequest, "Failed to find URL")
		return
	}

	c.Redirect(
		http.StatusPermanentRedirect,
		fullUrl,
	)
}

func main() {
	fmt.Println("Starting DB initialization")

	var err error
	database, err = storage.InitializeSQLiteDatabase("./tinyurl.db")
	if err != nil {
		log.Fatalf("Failed to initialize db: %v", err)
	}
	defer database.Close()

	fmt.Println("Succesfull DB initiazliation")

	r := gin.Default()
	r.GET("/:shortHash", redirectByShortLink)
	r.POST("/create", createShortLink)

	r.Run()
}
