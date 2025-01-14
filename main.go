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
	fmt.Print(url)

	urlPostfix := utils.ShortenUrl(url, 8)
	fmt.Printf("\nShorter url postfix %s for full url %s", urlPostfix, url)

	err := database.InsertURL(urlPostfix, url)
	if err != nil {
		responseWithError(c, http.StatusNotFound, "Failed to load URL from DB")
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"message": "URL recieved succesfully",
			"url":     url,
		},
	)
}

func redirectByShortLink(c *gin.Context) {
	shortHash := c.Param("shortHash")

	fullUrl, err := database.GetURL(shortHash)

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
	database, err = storage.InitializeSQLiteDatabase("./tinuurl.db")
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
