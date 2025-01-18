package main

import (
	"fmt"
	"log"
	"net/http"
	"tiny-url/storage"
	"tiny-url/utils"

	"github.com/gin-gonic/gin"
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
				"message":       "URL recieved succesfully",
				"url":           url,
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
			"message":       "URL recieved succesfully",
			"url":           url,
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

func getAllBindings(c *gin.Context) {
	bindings, err := database.GetAllBindings()

	if err != nil {
		responseWithError(c, http.StatusNotFound, "Failed to query bindings")
		return
	}

	c.JSON(
		http.StatusOK,
		bindings,
	)
}

func serveHome(c *gin.Context) {
	bindingsMap, err := database.GetAllBindings()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "base.html", gin.H{"bindings": nil})
		return
	}

	// Convert map[string]string to a slice of structs
	type Binding struct {
		FullURL  string
		ShortURL string
	}

	var bindings []Binding
	for fullURL, shortURL := range bindingsMap {
		bindings = append(bindings, Binding{FullURL: fullURL, ShortURL: shortURL})
	}

	// Debug: Print the bindings
	fmt.Printf("Bindings retrieved: %+v\n", bindings)

	c.HTML(
		http.StatusOK,
		"base.html",
		gin.H{
			"bindings": bindings,
		},
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

	r.LoadHTMLGlob("templates/*")

	r.GET("/:shortHash", redirectByShortLink)
	r.GET("/get_bindings", getAllBindings)
	r.POST("/create", createShortLink)

	r.GET("/", serveHome)

	r.Run()
}
