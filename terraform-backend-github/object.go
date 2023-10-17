package main

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v56/github"
)

type GithubObject struct {
	Owner string
	Repo  string
	Path  string
	Ref   string
}

func (obj *GithubObject) GetContent(ctx *gin.Context) (*github.RepositoryContent, bool, error) {
	_, resp, _ := client.Repositories.Get(ctx, obj.Owner, obj.Repo)
	if resp.StatusCode == 401 {
		return nil, false, errors.New("unauthorized")
	}
	if resp.StatusCode == 404 {
		return nil, false, errors.New("repo not found")
	}

	fileContent, _, resp, _ := client.Repositories.GetContents(ctx, obj.Owner, obj.Repo, obj.Path, &github.RepositoryContentGetOptions{
		Ref: *github.String(obj.Ref),
	})
	if resp.StatusCode == 401 {
		return nil, false, errors.New("unauthorized")
	}
	if resp.StatusCode == 404 {
		return nil, false, nil
	}

	return fileContent, true, nil
}

func NewGithubObject(ctx *gin.Context) (*GithubObject, error) {
	owner := ctx.Param("owner")
	repo := ctx.Param("repo")

	path := ctx.Param("path")
	if path == "/" {
		err := errors.New("path is required")
		return nil, err
	}
	path = path[1:]

	ref := ctx.Query("ref")
	if ref == "" {
		ref = "main"
	}

	return &GithubObject{
		Owner: owner,
		Repo:  repo,
		Path:  path,
		Ref:   ref,
	}, nil
}
