package segment

import (
	"fmt"
	"os"
	"parser/model"
	"parser/parser"
	"path/filepath"
	"time"
)

func SetSegmentIndex(entries []model.LogEntry) model.SegmentIndex {
	segmentIndex := model.SegmentIndex{
		ByLevel:     make(map[string][]int),
		ByComponent: make(map[string][]int),
		ByHost:      make(map[string][]int),
		ByReqID:     make(map[string][]int),
	}

	for i, entry := range entries {
		segmentIndex.ByLevel[string(entry.Level)] = append(segmentIndex.ByLevel[string(entry.Level)], i)
		segmentIndex.ByComponent[entry.Component] = append(segmentIndex.ByComponent[entry.Component], i)
		segmentIndex.ByHost[entry.Host] = append(segmentIndex.ByHost[entry.Host], i)
		segmentIndex.ByReqID[entry.Request_id] = append(segmentIndex.ByReqID[entry.Request_id], i)
	}

	return segmentIndex
}

func CreateSegments(path string) ([]model.Segment, error) {
	var Segments []model.Segment

	//enter folder
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory : %v", err)
	}

	//each file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var segment model.Segment
		segment.FileName = file.Name()
		path := filepath.Join(path, file.Name())
		allEntries, err := parser.ParseLogFile(path)
		if err != nil {
			fmt.Printf("Skipping file %s due to error: %v\n", path, err)
			continue

		}

		segment.LogEntries = allEntries

		for _, entry := range allEntries {
			if segment.StartTime.Equal(time.Time{}) || entry.Time.Before(segment.StartTime) {
				segment.StartTime = entry.Time
			}

			if entry.Time.After(segment.EndTime) {
				segment.EndTime = entry.Time
			}
		}
		segment.Index = SetSegmentIndex(allEntries)
		Segments = append(Segments, segment)
	}
	return Segments, nil
}
