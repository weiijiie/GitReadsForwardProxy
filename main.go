package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	gitHubUsername := os.Getenv("GH_USERNAME")
	if gitHubUsername == "" {
		log.Fatal("GH_USERNAME not set")
	}

	gitHubToken := os.Getenv("GH_TOKEN")
	if gitHubToken == "" {
		log.Fatal("GH_TOKEN not set")
	}

	router := gin.Default()
	gitHubClient := makeGitHubClient(gitHubUsername, gitHubToken)

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.Any("/github/*path", func(c *gin.Context) {
		path := c.Param("path")
		queryParams := make(url.Values)

		// We filter out all empty query params, as Golang encodes a query param
		// with key of `x` as `x=`, while the GitHub API expects `x`.
		for key := range c.Request.URL.Query() {
			val := c.Request.URL.Query().Get(key)
			if val == "" {
				continue
			}

			queryParams.Add(key, val)
		}

		log.Printf("Request Method: %s", c.Request.Method)
		log.Printf("Request Path: %s", path)
		log.Printf("Request Query Params: %v", queryParams)

		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Error reading request body: %+v", err)
		}

		resp, err := gitHubClient.R().
			SetBody(body).
			SetQueryParamsFromValues(queryParams).
			Execute(c.Request.Method, path)

		if err != nil {
			log.Printf("Error sending request: %+v, err: %+v", c.Request, err)
			return
		}

		contentType := resp.Header().Get("content-type")
		log.Printf("Response Content Type: %s", contentType)
		log.Printf("Response Status Code: %d", resp.StatusCode())

		// copy response headers
		for k, values := range resp.Header() {
			for _, val := range values {
				c.Writer.Header().Add(k, val)
			}
		}

		// return the response as-is
		c.Data(resp.StatusCode(), contentType, resp.Body())
	})

	log.Println("listening on", port)
	log.Fatal(router.Run(":" + port))
}

func makeGitHubClient(authUsername, authToken string) *resty.Client {
	client := resty.New().
		SetBaseURL("https://api.github.com").
		SetBasicAuth(authUsername, authToken).
		SetTimeout(30 * time.Second)

	return client
}
