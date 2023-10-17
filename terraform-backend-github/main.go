package main

import (
	"log"
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
	r.POST("/:owner/:repo/*path", post)
	r.Run(":8080")
}

func get(c *gin.Context) {
	state, err := NewTerraformState(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, exists, err := state.Content(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(404, gin.H{
			"error": "file not found",
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

func post(c *gin.Context) {
	state, err := NewTerraformState(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, exists, err := state.Content(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	if exists {
		_, _, err = gh.Repositories.UpdateFile(c, state.Owner, state.Repo, state.Path, &github.RepositoryContentFileOptions{
			SHA:     fileContent.SHA,
			Message: github.String("update file"),
			Content: body,
			Branch:  github.String(state.Ref),
		})

		if err != nil {
			log.Println(err)
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Data(200, "application/json", []byte("{}"))
	} else {
		_, _, err = gh.Repositories.CreateFile(c, state.Owner, state.Repo, state.Path, &github.RepositoryContentFileOptions{
			Message: github.String("create file"),
			Content: body,
			Branch:  github.String(state.Ref),
		})

		if err != nil {
			log.Println(err)
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Data(200, "application/json", []byte("{}"))
		return
	}
}
