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
	owner := c.Param("owner")
	repoName := c.Param("repo")
	path := c.Param("path")
	if path == "/" {
		c.JSON(400, gin.H{
			"error": "path is required",
		})
	}
	path = path[1:]
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

	fileContent, _, resp, err := gh.Repositories.GetContents(c, owner, repoName, path, &github.RepositoryContentGetOptions{
		Ref: ref,
	})

	if resp.StatusCode == 404 {
		c.JSON(404, gin.H{
			"error": "not found",
		})
		return
	}

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

func post(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	path := c.Param("path")
	if path == "/" {
		c.JSON(400, gin.H{
			"error": "path is required",
		})
	}
	path = path[1:]
	ref := c.Query("ref")
	if ref == "" {
		ref = "main"
	}

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	// check if file exists
	fileContent, _, resp, _ := gh.Repositories.GetContents(c, owner, repoName, path, &github.RepositoryContentGetOptions{
		Ref: ref,
	})

	if resp.StatusCode == 200 {
		// update the file in the repo
		_, _, err = gh.Repositories.UpdateFile(c, owner, repoName, path, &github.RepositoryContentFileOptions{
			SHA:     fileContent.SHA,
			Message: github.String("update file"),
			Content: body,
			Branch:  github.String(ref),
		})

		if err != nil {
			log.Println(err)
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Data(200, "application/json", []byte("{}"))
	} else if resp.StatusCode == 404 {
		// create a new file in the repo
		_, _, err = gh.Repositories.CreateFile(c, owner, repoName, path, &github.RepositoryContentFileOptions{
			Message: github.String("create file"),
			Content: body,
			Branch:  github.String(ref),
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
