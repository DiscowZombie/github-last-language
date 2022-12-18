package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
	"net/http"
	"os"
)

const WebsiteTitle = "Github Last Language"

type RepositoryLanguage struct {
	Repository *github.Repository
	Loc        int
}

func handleMain(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl.html", gin.H{
		"title": WebsiteTitle,
	})

	// TODO retrieve (and propose) all github langs?
}

// TODO context par value or ptr?
func handleSearch(ctx *context.Context, ghClient *github.Client, c *gin.Context) {
	languageQuery := c.Query("language")
	if languageQuery == "" {
		c.Redirect(http.StatusUnprocessableEntity, "/")
		return
	}

	// Get repositories
	search := &github.SearchOptions{
		Sort:  "updated",
		Order: "desc",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 10, // TODO
		},
	}

	repositories, resp, err := ghClient.Search.Repositories(*ctx, "archived:false", search)
	// Check request response
	if err != nil || resp.StatusCode != http.StatusOK {
		c.Status(http.StatusInternalServerError)
		return
	}

	// TODO uppercase first letter
	requestedLang := languageQuery

	var repos []RepositoryLanguage

	for _, repo := range repositories.Repositories {
		languages, _, err := ghClient.Repositories.ListLanguages(*ctx, repo.GetOwner().GetLogin(), repo.GetName())
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}

		// Lines of codes for the request lang
		loc := languages[requestedLang]

		if loc > 0 {
			repos = append(repos, RepositoryLanguage{Repository: repo, Loc: loc})
		}
	}

	c.HTML(http.StatusOK, "search.tmpl.html", gin.H{
		"title": WebsiteTitle,
		"repos": repos,
	})
}

func main() {
	ctx := context.Background()

	var client *http.Client

	env, envSet := os.LookupEnv("GITHUB_TOKEN")
	if envSet {
		// Auth to GitHub with a personal access token
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: env})
		client = oauth2.NewClient(ctx, ts)
	}

	// Github client
	ghClient := github.NewClient(client)

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
