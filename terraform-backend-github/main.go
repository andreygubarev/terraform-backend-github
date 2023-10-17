package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v56/github"
)

var client *github.Client

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		panic("GITHUB_TOKEN is required")
	}
	client = github.NewClient(nil).WithAuthToken(token)

	r := gin.Default()
	r.GET("/:owner/:repo/*path", getHandler)
	r.POST("/:owner/:repo/*path", postHandler)
	r.DELETE("/:owner/:repo/*path", deleteHandler)
	r.Handle("LOCK", "/:owner/:repo/*path", lockHandler)
	r.Handle("UNLOCK", "/:owner/:repo/*path", deleteHandler)
	r.Run(":8080")
}

func getHandler(c *gin.Context) {
	obj, err := NewGithubObject(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, exists, err := obj.GetContent(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "file not found",
		})
		return
	}

	content, err := fileContent.GetContent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Data(200, "application/json", []byte(content))
}

func postHandler(c *gin.Context) {
	obj, err := NewGithubObject(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, exists, err := obj.GetContent(c)
	if err != nil {
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
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
		_, _, err = client.Repositories.UpdateFile(c, obj.Owner, obj.Repo, obj.Path, &github.RepositoryContentFileOptions{
			SHA:     fileContent.SHA,
			Message: github.String("update terraform state"),
			Content: body,
			Branch:  github.String(obj.Ref),
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.Data(http.StatusOK, "application/json", []byte("{}"))
		return
	} else {
		_, _, err = client.Repositories.CreateFile(c, obj.Owner, obj.Repo, obj.Path, &github.RepositoryContentFileOptions{
			Message: github.String("create terraform state"),
			Content: body,
			Branch:  github.String(obj.Ref),
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Data(http.StatusOK, "application/json", []byte("{}"))
		return
	}
}

func deleteHandler(c *gin.Context) {
	obj, err := NewGithubObject(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, exists, err := obj.GetContent(c)
	if err != nil {
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "file not found",
		})
		return
	}

	_, _, err = client.Repositories.DeleteFile(c, obj.Owner, obj.Repo, obj.Path, &github.RepositoryContentFileOptions{
		SHA:     fileContent.SHA,
		Message: github.String("delete terraform state"),
		Branch:  github.String(obj.Ref),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.Data(http.StatusOK, "application/json", []byte("{}"))
}

func lockHandler(c *gin.Context) {
	obj, err := NewGithubObject(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, exists, err := obj.GetContent(c)
	if err != nil {
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
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
		_, _, err = client.Repositories.CreateFile(c, obj.Owner, obj.Repo, obj.Path, &github.RepositoryContentFileOptions{
			Message: github.String("create terraform state lock"),
			Content: body,
			Branch:  github.String(obj.Ref),
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Data(http.StatusOK, "application/json", body)
		return
	}
}
