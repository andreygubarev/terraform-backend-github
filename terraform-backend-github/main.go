package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v56/github"
)

var gh *github.Client

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		panic("GITHUB_TOKEN is required")
	}
	gh = github.NewClient(nil).WithAuthToken(token)

	r := gin.Default()
	r.GET("/:owner/:repo/*path", getHandler)
	r.POST("/:owner/:repo/*path", postHandler)
	r.DELETE("/:owner/:repo/*path", deleteHandler)
	r.Handle("LOCK", "/:owner/:repo/*path", lockHandler)
	r.Handle("UNLOCK", "/:owner/:repo/*path", deleteHandler)
	r.Run(":8080")
}

func getHandler(c *gin.Context) {
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

func postHandler(c *gin.Context) {
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
			Message: github.String("update terraform state"),
			Content: body,
			Branch:  github.String(state.Ref),
		})

		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Data(200, "application/json", []byte("{}"))
		return
	} else {
		_, _, err = gh.Repositories.CreateFile(c, state.Owner, state.Repo, state.Path, &github.RepositoryContentFileOptions{
			Message: github.String("create terraform state"),
			Content: body,
			Branch:  github.String(state.Ref),
		})

		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Data(200, "application/json", []byte("{}"))
		return
	}
}

func deleteHandler(c *gin.Context) {
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

	_, _, err = gh.Repositories.DeleteFile(c, state.Owner, state.Repo, state.Path, &github.RepositoryContentFileOptions{
		SHA:     fileContent.SHA,
		Message: github.String("delete terraform state"),
		Branch:  github.String(state.Ref),
	})

	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.Data(200, "application/json", []byte("{}"))
}

func lockHandler(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if exists {
		content, err := fileContent.GetContent()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Data(http.StatusConflict, "application/json", []byte(content))
		return
	} else {
		_, _, err = gh.Repositories.CreateFile(c, state.Owner, state.Repo, state.Path, &github.RepositoryContentFileOptions{
			Message: github.String("create terraform state lock"),
			Content: body,
			Branch:  github.String(state.Ref),
		})

		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Data(200, "application/json", body)
		return
	}
}
