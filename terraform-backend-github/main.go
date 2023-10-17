package main

import (
	"errors"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v56/github"
)

var gh *github.Client

type TerraformState struct {
	Owner string
	Repo  string
	Path  string
	Ref   string
}

func NewTerraformState(c *gin.Context) (*TerraformState, error) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	path := c.Param("path")
	if path == "/" {
		err := errors.New("path is required")
		return nil, err
	}
	path = path[1:]

	ref := c.Query("ref")
	if ref == "" {
		ref = "main"
	}

	return &TerraformState{
		Owner: owner,
		Repo:  repo,
		Path:  path,
		Ref:   ref,
	}, nil
}

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

	_, _, err = gh.Repositories.Get(c, state.Owner, state.Repo)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, _, resp, err := gh.Repositories.GetContents(c, state.Owner, state.Repo, state.Path, &github.RepositoryContentGetOptions{
		Ref: state.Ref,
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
