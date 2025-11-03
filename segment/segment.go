package segment

import (
	"fmt"
	"os"
	"parser/model"
	"parser/parser"
	"path/filepath"
	"sync"
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
	start := time.Now()
	var Segments []model.Segment

	var wg sync.WaitGroup
	var mu sync.Mutex

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

		wg.Add(1)

		go func(file os.DirEntry) {
			defer wg.Done()

			var segment model.Segment
			segment.FileName = file.Name()
			filPath := filepath.Join(path, file.Name())
			allEntries, err := parser.ParseLogFile(filPath)
			if err != nil {
				fmt.Printf("Skipping file %s due to error: %v\n", path, err)
				return

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

			// lock before writing to shared slice
			mu.Lock()
			Segments = append(Segments, segment)
			mu.Unlock()

		}(file)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println("Segmenting duration:", elapsed)
	return Segments, nil
}
