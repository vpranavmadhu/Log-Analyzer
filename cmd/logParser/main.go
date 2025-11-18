package main

import (
	"fmt"
	"log"
	"os"
	models "parser/pkg/dbmodels"
	"parser/pkg/parser"
	"path/filepath"
)

const dbUrl = "postgresql:///logAnalyzerDB?host=/var/run/postgresql/"

func handleCommand(args []string) error {
	db, err := models.CreateDB(dbUrl)
	if err != nil {
		return err
	}
	switch args[0] {
	case "init":
		err := models.InitDb(db)
		if err != nil {
			return err
		}
	case "add":
		folderName := args[1] //folder name
		fmt.Printf("Adding logs from %s\n", folderName)

		files, err := os.ReadDir(folderName)
		if err != nil {
			return fmt.Errorf("failed to read directory : %v", err)
		}

		for _, file := range files { //each files
			if file.IsDir() {
				continue
			}
			fmt.Println("Entered file: ", file.Name())
			path := filepath.Join(folderName, file.Name())
			entries, err := parser.ParseLogFile(path)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				models.AddEntry(db, entry)
			}
		}

		return nil
	case "query":
		// query := strings.Join(args[1:], " ")
		queryList := args[1:]

		entries, err := models.Query(db, queryList)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			fmt.Println(entry)
		}
		fmt.Printf("%d entries matched: \n", len(entries))
		return nil

	case "web":
		r := SetupRoutes(db)
		log.Println("Server running at http://localhost:8080")
		return r.Run(":8080")

	default:
		return fmt.Errorf("unknown command: %s (expected: init | add | query)", args[0])
	}
	return nil
}

// func runserver() error {
// 	db, err := models.CreateDB(dbUrl)
// 	if err != nil {
// 		return err
// 	}

// 	r := gin.Default()

// 	r.LoadHTMLGlob("templates/*.html")

// 	r.GET("/logs", func(c *gin.Context) {

// 		parts := []string{}
// 		for key, vals := range c.Request.URL.Query() {
// 			parts = append(parts, fmt.Sprintf("%s=%s", key, vals[0]))
// 		}

// 		entries, err := models.Query(db, parts)
// 		if err != nil {
// 			// c.JSON(400, gin.H{"error": err.Error()})
// 			c.HTML(500, "result.html", gin.H{
// 				"error": err.Error(),
// 			})
// 			return
// 		}

// 		// c.JSON(200, entries)
// 		c.HTML(200, "result.html", entries)
// 	})

// 	// fmt.Println("Web server running at http://localhost:8080/logs")
// 	return r.Run(":8080")
// }

func main() {
	err := handleCommand(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in invocation %v", err)
		os.Exit(-1)
	}
}
