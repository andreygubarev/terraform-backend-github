package main

import (
	"errors"

	"github.com/gin-gonic/gin"
)

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
