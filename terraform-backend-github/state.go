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

func (g *GithubObject) GetContent(c *gin.Context) (*github.RepositoryContent, bool, error) {
	_, resp, _ := client.Repositories.Get(c, g.Owner, g.Repo)
	if resp.StatusCode == 401 {
		return nil, false, errors.New("unauthorized")
	}
	if resp.StatusCode == 404 {
		return nil, false, errors.New("repo not found")
	}

	fileContent, _, resp, _ := client.Repositories.GetContents(c, g.Owner, g.Repo, g.Path, &github.RepositoryContentGetOptions{
		Ref: *github.String(g.Ref),
	})
	if resp.StatusCode == 401 {
		return nil, false, errors.New("unauthorized")
	}
	if resp.StatusCode == 404 {
		return nil, false, nil
	}

	return fileContent, true, nil
}

func NewGithubObject(c *gin.Context) (*GithubObject, error) {
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

	return &GithubObject{
		Owner: owner,
		Repo:  repo,
		Path:  path,
		Ref:   ref,
	}, nil
}
