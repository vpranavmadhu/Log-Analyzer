package main

import (
	"fmt"
	"net/http"
	models "parser/pkg/dbmodels"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// logs/
func showAllLogs(c *gin.Context) {
	entries, err := models.Query(DB, []string{}) //empty filter to get all logs
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"entries": entries,
		"count":   len(entries),
	})
}

func showAllLogsPaginated(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))

	offset := page * pageSize

	var entries []models.Entry
	var count int64

	DB.Model(&models.Entry{}).Count(&count)
	err := DB.
		Model(&models.Entry{}).
		Preload("Level").
		Preload("Component").
		Preload("Host").
		Order("id ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&entries).Error

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"count":   count,
	})
	fmt.Println("page:", page, "pageSize:", pageSize, "offset:", offset)

}

func showAllLogsHTML(c *gin.Context) {
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

// filteredlogs/
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

func filterLogsJSON(c *gin.Context) {
	var body map[string][]string

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid JSON body",
		})
		return
	}

	queryParts := []string{}

	for key, vals := range body {
		if len(vals) > 0 {
			joined := strings.Join(vals, ",")
			queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, joined))
		}
	}

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

func filterLogsPaginated(c *gin.Context) {

	var body map[string][]string

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid JSON body",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "100"))

	offset := page * pageSize

	queryParts := []string{}

	for key, vals := range body {
		if len(vals) > 0 {
			joined := strings.Join(vals, ",")
			queryParts = append(queryParts, fmt.Sprintf("%s=%s", key, joined))
		}
	}

	filteredEntries, err := models.Query(DB, queryParts)
	if err != nil {
		c.JSON(500, err)
		return
	}

	total := len(filteredEntries)
	// Apply manual pagination to filtered data
	start := offset
	end := offset + pageSize

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	pageEntries := filteredEntries[start:end]

	c.JSON(http.StatusOK, gin.H{
		"entries": pageEntries,
		"count":   total,
	})

}
