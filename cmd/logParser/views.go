package main

import (
	"fmt"
	models "parser/pkg/dbmodels"
	"strings"

	"github.com/gin-gonic/gin"
)

func showAllLogs(c *gin.Context) {
	entries, err := models.Query(DB, []string{}) //empty filter to get all logs
	if err != nil {
		c.HTML(500, "result.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(200, "result.html", gin.H{
		"entries": entries,
		"count":   len(entries),
	})
}

func filterLogs(c *gin.Context) {
	queryParts := []string{}

	c.Request.ParseForm()
	formData := c.Request.PostForm

	result := make(map[string][]string)

	for key, values := range formData {
		if len(values) > 0 && values[0] != "" {
			// result[key] = strings.Split(values[0], ",")
			result[key] = values
		}
	}

	// Build filters
	// for key, vals := range result {
	// 	if len(vals) > 0 {
	// 		queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, vals[0]))
	// 	}
	// }

	for key, vals := range result {
		if len(vals) > 0 {
			// join multiple values into one comma string: "INFO,DEBUG"
			joined := strings.Join(vals, ",")
			queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, joined))
		}
	}

	// Execute query
	entries, err := models.Query(DB, queryParts)
	if err != nil {
		c.JSON(500, err)
		return
	}
	c.JSON(200, gin.H{
		"entries": entries,
		"count":   len(entries),
	})
}
