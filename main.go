package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"tiny-url/utils"
)

var urlStorage = make(map[string]string)

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
	fmt.Printf("\nShorter url postfix %s", urlPostfix)

	urlStorage[urlPostfix] = url
	fmt.Printf("\nStorage %s", urlStorage)

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

	fullUrl, fullUrlExists := urlStorage[shortHash]

	if !fullUrlExists {
		responseWithError(c, http.StatusBadRequest, "Failed to find URL")
		return
	}

	c.Redirect(
		http.StatusPermanentRedirect,
		fullUrl,
	)
}

func main() {
	fmt.Print("Hello world!")

	r := gin.Default()
	r.GET("/:shortHash", redirectByShortLink)
	r.POST("/create", createShortLink)

	r.Run()
}
