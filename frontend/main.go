package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const (
	predifinedUsername = "vadbeg"
	predifinedPassword = "123456"
)

// Generate random session key once when binary starts
var globalSessionKey = fmt.Sprintf("session_%d_%d", time.Now().UnixNano(), rand.Int63())

type LinkResponse struct {
	FullURL      string `json:"full_url"`
	ShortPostfix string `json:"short_postfix"`
}

type Binding struct {
	FullURL  string
	ShortURL string
}

func getAllBindings() ([]Binding, error) {
	resp, err := http.Get("http://0.0.0.0:8080/get_bindings")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch bindings: " + resp.Status)
	}

	var bindingsMap map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&bindingsMap); err != nil {
		return nil, err
	}

	var bindings []Binding
	for fullURL, shortURL := range bindingsMap {
		bindings = append(bindings, Binding{FullURL: fullURL, ShortURL: shortURL})
	}

	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].FullURL < bindings[j].FullURL
	})

	return bindings, nil
}

func serveHome(c *gin.Context) {
	bindings, err := getAllBindings()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	log.Printf("Bindings retrieved: %+v\n", bindings)

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
	body, err := io.ReadAll(resp.Body)
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
	log.Println("Full URL:", linkResponse.FullURL)
	log.Println("Short Postfix:", linkResponse.ShortPostfix)

	c.Redirect(
		http.StatusPermanentRedirect,
		linkResponse.FullURL,
	)
}

func createShortLink(c *gin.Context) {
	fullURL := c.PostForm("URL")
	log.Println(fullURL)

	requestUrl := "http://0.0.0.0:8080/create"

	values := map[string]string{"URL": fullURL}
	jsonValue, _ := json.Marshal(values)
	resp, _ := http.Post(requestUrl, "application/json", bytes.NewBuffer(jsonValue))

	log.Printf("Tried creating shortened url from %s, got %d status code", fullURL, resp.StatusCode)

	bindings, err := getAllBindings()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	c.HTML(
		http.StatusOK,
		"chatbox.html",
		gin.H{
			"bindings": bindings,
		},
	)
}

func removeLink(c *gin.Context) {
	hash := c.Param("shortHash")

	requestUrl := "http://0.0.0.0:8080/remove/" + hash

	resp, _ := http.Post(requestUrl, "application/json", nil)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to remove link: %d", resp.StatusCode)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to remove link"})
		return
	}

	bindings, err := getAllBindings()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	c.HTML(
		http.StatusOK,
		"chatbox.html",
		gin.H{
			"bindings": bindings,
		},
	)
}

func authRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(globalSessionKey)

	if user == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
		return
	}

	c.Next()
}

func loginHandler(c *gin.Context) {
	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "login.html", gin.H{})
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")

	log.Println(predifinedUsername, username)
	log.Println(predifinedPassword, password)

	if username == predifinedUsername && password == predifinedPassword {
		session := sessions.Default(c)

		session.Set(globalSessionKey, username)
		session.Save()
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	log.Println("you are fucking here")

	c.String(http.StatusOK, "<div class='error'>Invalid credentials</div>")
}

func main() {
	r := gin.Default()

	store := cookie.NewStore([]byte("secret_key"))
	r.Use(sessions.Sessions(globalSessionKey, store))

	r.LoadHTMLGlob("templates/*")

	r.GET("/login", loginHandler)
	r.POST("/login", loginHandler)

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	protected := r.Group("/")
	protected.Use(authRequired)
	{
		protected.GET("/", serveHome)
		protected.GET("/:shortHash", redirectByShortLink)
		protected.POST("/create", createShortLink)
		protected.DELETE("/remove/:shortHash", removeLink)
	}

	r.Run("0.0.0.0:8081")
}
