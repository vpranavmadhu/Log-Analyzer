package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"parser/filter"
	"parser/segment"
	"strings"
	"time"
)

func split(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func parseTime(value string, label string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	parsed, err := time.Parse("2006-01-02 15:04:05", value)
	if err != nil {
		slog.Error("Error parsing time", "field", label, "input", value, "error", err)
		return time.Time{}, err
	}
	return parsed, nil
}

func main() {

	start := time.Now()
	path := flag.String("path", "/home/pranavmadhu/learn/LogAnalyzer/logs", "Path to logs directory")
	level := flag.String("level", "", "Filter by log level")
	component := flag.String("component", "", "Filter by component")
	host := flag.String("host", "", "Filter by host")
	reqID := flag.String("reqID", "", "Filter by reqID")
	startTimeString := flag.String("after", "", "Filter by start time")
	endTimeString := flag.String("before", "", "Filter by end time")
	flag.Parse()

	segments, err := segment.CreateSegments(*path)

	if err != nil {
		slog.Error("Error in creating segments: ", "Error", err)
		os.Exit(1)
	}

	levels := split(*level)
	components := split(*component)
	hosts := split(*host)
	reqIDs := split(*reqID)

	startTime, _ := parseTime(*startTimeString, "after")
	endTime, _ := parseTime(*endTimeString, "before")

	filteredLogs := filter.FilterEntries(segments, levels, components, hosts, reqIDs, startTime, endTime)

	for _, entry := range filteredLogs {
		fmt.Println(entry.Log)
	}
	fmt.Printf("Found %d matching entries\n", len(filteredLogs))

	elapsed := time.Since(start)
	fmt.Printf("duration: %s\n", elapsed)
}
