package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"net/http"
	"os"
	"sort"
)

const WebsiteTitle = "Github Last Language"

type RepositoryLanguage struct {
	Name          string
	NameWithOwner string
	Url           string
	// GitHub returns the number of bytes of code, not exactly the number of lines
	Loc int
}

func handleMain(c *gin.Context) {
	// Unfortunately GitHub does not provide any endpoint to retrieve a list of languages
	c.HTML(http.StatusOK, "index.tmpl.html", gin.H{
		"title": WebsiteTitle,
	})
}

// TODO context par value or ptr?
func handleSearch(ctx *context.Context, ghClient *githubv4.Client, c *gin.Context) {
	languageQuery := c.Query("language")
	if languageQuery == "" {
		// Missing required parameter
		c.Status(http.StatusBadRequest)
		return
	}

	// TODO Update request
	// Make the request
	var q struct {
		Repository struct {
			RepositoryCount int
			Edges           []struct {
				Node struct {
					Repository struct {
						Name          string
						NameWithOwner string
						Url           string
						Languages     struct {
							TotalCount int
							Edges      []struct {
								Size int
								Node struct {
									Name string
								}
							}
						} `graphql:"languages(first: 100)"`
					} `graphql:"... on Repository"`
				}
			}
		} `graphql:"search(query: \"is:public sort:author-date-desc\", type: REPOSITORY, first: 100)"`
	}

	err := ghClient.Query(*ctx, &q, nil)

	// Check request response
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	var repos []RepositoryLanguage

	// Iterate each repository
	for _, edge := range q.Repository.Edges {
		loc := 0

		// Find the LOC for the requested language on the repository
		for _, s := range edge.Node.Repository.Languages.Edges {
			if s.Node.Name == languageQuery {
				loc = s.Size
				break
			}
		}

		// The repository contain at least one line of code for the requested language
		if loc > 0 {
			repos = append(repos, RepositoryLanguage{
				Name:          edge.Node.Repository.Name,
				NameWithOwner: edge.Node.Repository.NameWithOwner,
				Url:           edge.Node.Repository.Url,
				Loc:           loc,
			})
		}
	}

	// Sort the returned list of repositories
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Loc > repos[j].Loc // DESC
	})

	c.HTML(http.StatusOK, "search.tmpl.html", gin.H{
		"title": WebsiteTitle,
		"repos": repos,
	})
}

func main() {
	ctx := context.Background()

	var client *http.Client

	env, envSet := os.LookupEnv("GITHUB_TOKEN")
	if !envSet {
		/*
			Generate a token from https://github.com/settings/tokens with the "public_repo" scope.
			"Fine-grained tokens" are not supported to access the GitHub GraphQL API for now.
		*/
		fmt.Printf("The env variable \"GITHUB_TOKEN\" should contain a classic token obtained from https://github.com/settings/tokens with the \"public_repo\" scope.")
		os.Exit(1)
		return
	}

	// Auth to GitHub with a personal access token
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: env})
	client = oauth2.NewClient(ctx, ts)

	// Github client
	ghClient := githubv4.NewClient(client)

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
		fmt.Printf("unable to start web-server: %v", err)
		os.Exit(2)
	}
}
