package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

func SetupRoutes(db *gorm.DB) *gin.Engine {
	DB = db

	r := gin.Default()

	r.Use(CORSMiddleware())
	r.LoadHTMLGlob("templates/*.html")

	r.GET("/", showAllLogsHTML) //for html
	r.GET("/logs", showAllLogs) //first time loading
	r.GET("/alllogs", showAllLogsPaginated)
	r.POST("/filteredlogs", filterLogs)   // after searching
	r.POST("/filterlogs", filterLogsJSON) //react filter
	r.POST("/filter", filterLogsPaginated)

	return r
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
