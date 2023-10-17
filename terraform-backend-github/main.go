package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v56/github"
)

var gh *github.Client

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	gh = github.NewClient(nil).WithAuthToken(token)

	r := gin.Default()
	r.GET("/:owner/:repo/*path", get)
	r.Run(":8080")
}

func get(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	path := c.Param("path")

	if path == "/" {
		c.JSON(400, gin.H{
			"error": "path is required",
		})
	}

	ref := c.Query("ref")
	if ref == "" {
		ref = "main"
	}

	_, _, err := gh.Repositories.Get(c, owner, repoName)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, _, _, err := gh.Repositories.GetContents(c, owner, repoName, path, &github.RepositoryContentGetOptions{
		Ref: ref,
	})

	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	content, err := fileContent.GetContent()
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Data(200, "application/json", []byte(content))
}
