package main

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v56/github"
)

type TerraformState struct {
	Owner string
	Repo  string
	Path  string
	Ref   string
}

func (t *TerraformState) Content(c *gin.Context) (*github.RepositoryContent, bool, error) {
	_, resp, _ := gh.Repositories.Get(c, t.Owner, t.Repo)
	if resp.StatusCode == 401 {
		return nil, false, errors.New("unauthorized")
	}
	if resp.StatusCode == 404 {
		return nil, false, errors.New("repo not found")
	}

	fileContent, _, resp, _ := gh.Repositories.GetContents(c, t.Owner, t.Repo, t.Path, &github.RepositoryContentGetOptions{
		Ref: *github.String(t.Ref),
	})
	if resp.StatusCode == 401 {
		return nil, false, errors.New("unauthorized")
	}
	if resp.StatusCode == 404 {
		return nil, false, nil
	}

	return fileContent, true, nil
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
