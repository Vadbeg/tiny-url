package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LinkResponse struct {
	FullURL      string `json:"full_url"`
	ShortPostfix string `json:"short_postfix"`
}

func responseWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(
		statusCode,
		gin.H{
			"error": message,
		},
	)
}

func serveHome(c *gin.Context) {
	resp, _ := http.Get("http://0.0.0.0:8080/get_bindings")

	var bindingsMap map[string]string
	json.NewDecoder(resp.Body).Decode(&bindingsMap)

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

func redirectByShortLink(c *gin.Context) {
	shortHash := c.Param("shortHash")

	// Construct the full URL with the parameter
	fullURL := "http://0.0.0.0:8080/get_url/" + shortHash

	// Send the GET request
	resp, err := http.Get(fullURL)
	if err != nil {
		log.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to get data. Status Code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Print the raw response body (for debugging)
	fmt.Println("Response Body:", string(body))

	// Parse the JSON response into a struct
	var linkResponse LinkResponse
	err = json.Unmarshal(body, &linkResponse)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Print the parsed response
	fmt.Println("Full URL:", linkResponse.FullURL)
	fmt.Println("Short Postfix:", linkResponse.ShortPostfix)

	c.Redirect(
		http.StatusPermanentRedirect,
		linkResponse.FullURL,
	)
}

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.GET("/", serveHome)
	r.GET("/:shortHash", redirectByShortLink)

	r.Run("0.0.0.0:8081")
}
