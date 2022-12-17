package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v48/github"
	"net/http"
)

const WebsiteTitle = "Github Last Language"

func handleMain(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl.html", gin.H{
		"title": WebsiteTitle,
	})
}

// TODO context par value or ptr?
func handleSearch(ctx *context.Context, ghClient *github.Client, c *gin.Context) {
	// Get repositories
	search := &github.SearchOptions{
		Sort:  "updated",
		Order: "desc",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	repositories, resp, err := ghClient.Search.Repositories(*ctx, "archived:false", search)
	// Check request response
	if err != nil || resp.StatusCode != http.StatusOK {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Filter matched repositories
	var matched []*github.Repository
	for _, repo := range repositories.Repositories {
		if repo.GetLanguage() == "Java" {
			matched = append(matched, repo)
		}
	}

	for _, repository := range matched {
		fmt.Printf("URL: %s\n", *repository.URL)
	}

	c.Status(http.StatusOK)
}

func main() {
	// Github client
	ghClient := github.NewClient(nil)
	ctx := context.Background()

	// Router
	r := gin.Default()

	// Load html templates
	r.LoadHTMLGlob("templates/*")

	// Routes
	r.GET("/", handleMain)
	r.GET("/search", func(c *gin.Context) {
		handleSearch(&ctx, ghClient, c)
	})

	// Start the web-server on 0.0.0.0:8080
	err := r.Run()
	if err != nil {
		_ = fmt.Errorf("unable to start web-server: %v", err)
	}
}
