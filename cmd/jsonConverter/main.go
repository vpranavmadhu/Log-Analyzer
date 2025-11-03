package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"parser/segment"
)

func main() {

	inputFile := flag.String("in", "/home/pranavmadhu/learn/LogAnalyzer/logs", "Path to the log directory")
	outputFile := flag.String("out", "logs.json", "Output file path")
	flag.Parse()

	segments, err := segment.CreateSegments(*inputFile)
	if err != nil {
		slog.Error("Cannot create segments", "error", err)
	}

	file, err := os.Create(*outputFile)
	if err != nil {
		slog.Error("Cannot create file", "error", err)
	}
	defer file.Close()
	jsonData, err := json.MarshalIndent(segments, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Printf(" JSON data successfully written to file : %v\n", *outputFile)
}
