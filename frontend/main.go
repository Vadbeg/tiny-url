package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func serveHome(c *gin.Context) {
	bindingsMap := make(map[string]string)
	bindingsMap["test"] = "test"

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
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.GET("/", serveHome)

	r.Run()
}
