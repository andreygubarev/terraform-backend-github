package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.Any("/:org/:repo/*path", get)
	r.Run(":8080")
}

func get(c *gin.Context) {
	org := c.Param("org")
	repo := c.Param("repo")
	path := c.Param("path")

	if path == "/" {
		c.JSON(400, gin.H{
			"error": "path is required",
		})
	}

	ref := c.Query("ref")
	if ref == "" {
		ref = "main"
	}

	c.JSON(200, gin.H{
		"org":  org,
		"repo": repo,
		"path": path,
	})
}
