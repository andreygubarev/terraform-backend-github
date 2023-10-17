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
	r.GET("/:owner/:repo/*path", ReadHandler)
	r.POST("/:owner/:repo/*path", CreateHandler)
	r.DELETE("/:owner/:repo/*path", DeleteHandler)
	r.Handle("LOCK", "/:owner/:repo/*path", LockHandler)
	r.Handle("UNLOCK", "/:owner/:repo/*path", DeleteHandler)
	r.Run(":8080")
}

func ReadHandler(ctx *gin.Context) {
	obj, err := NewGithubObject(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, fileExists, err := obj.GetContent(ctx)
	if err != nil {
		if err.Error() == "unauthorized" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	if !fileExists {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "file not found",
		})
		return
	}

	content, err := fileContent.GetContent()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Data(200, "application/json", []byte(content))
}

func CreateHandler(ctx *gin.Context) {
	obj, err := NewGithubObject(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, fileExists, err := obj.GetContent(ctx)
	if err != nil {
		if err.Error() == "unauthorized" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	body, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if fileExists {
		opts := &github.RepositoryContentFileOptions{
			Branch:  github.String(obj.Ref),
			Content: body,
			Message: github.String("update terraform state"),
			SHA:     fileContent.SHA,
		}
		_, _, err = client.Repositories.UpdateFile(ctx, obj.Owner, obj.Repo, obj.Path, opts)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		ctx.Data(http.StatusOK, "application/json", []byte("{}"))
		return
	} else {
		_, _, err = client.Repositories.CreateFile(ctx, obj.Owner, obj.Repo, obj.Path, &github.RepositoryContentFileOptions{
			Branch:  github.String(obj.Ref),
			Content: body,
			Message: github.String("Create a Terraform state"),
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.Data(http.StatusOK, "application/json", []byte("{}"))
		return
	}
}

func DeleteHandler(ctx *gin.Context) {
	obj, err := NewGithubObject(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, fileExists, err := obj.GetContent(ctx)
	if err != nil {
		if err.Error() == "unauthorized" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	if !fileExists {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "file not found",
		})
		return
	}

	opts := &github.RepositoryContentFileOptions{
		Branch:  github.String(obj.Ref),
		Message: github.String("Delete the Terraform state"),
		SHA:     fileContent.SHA,
	}
	_, _, err = client.Repositories.DeleteFile(ctx, obj.Owner, obj.Repo, obj.Path, opts)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Data(http.StatusOK, "application/json", []byte("{}"))
}

func LockHandler(ctx *gin.Context) {
	obj, err := NewGithubObject(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	fileContent, fileExists, err := obj.GetContent(ctx)
	if err != nil {
		if err.Error() == "unauthorized" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	body, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if fileExists {
		content, err := fileContent.GetContent()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.Data(http.StatusConflict, "application/json", []byte(content))
		return
	} else {
		opts := &github.RepositoryContentFileOptions{
			Branch:  github.String(obj.Ref),
			Content: body,
			Message: github.String("Create a Terraform state lock"),
		}
		_, _, err = client.Repositories.CreateFile(ctx, obj.Owner, obj.Repo, obj.Path, opts)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.Data(http.StatusOK, "application/json", body)
		return
	}
}
