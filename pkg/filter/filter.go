package filter

import (
	"fmt"
	"parser/model"
	"sync"
	"time"
)

func FilterEntries(
	segments []model.Segment,
	levels, components, hosts, reqIDs []string,
	startTime, endTime time.Time,
) []model.LogEntry {

	start := time.Now()
	var result []model.LogEntry
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, segment := range segments {
		wg.Add(1)
		go func(segment model.Segment) {
			defer wg.Done()

			//skip segment outside time range
			if !startTime.IsZero() && segment.EndTime.Before(startTime) {
				return
			}
			if !endTime.IsZero() && segment.StartTime.After(endTime) {
				return
			}

			totalFilters := 0
			matchedIndex := make(map[int]bool)

			// by level
			if len(levels) > 0 {
				totalFilters++
				for _, level := range levels {
					for _, idx := range segment.Index.ByLevel[level] {
						matchedIndex[idx] = true
					}
				}
			}
			// by component
			if len(components) > 0 {
				totalFilters++
				componentFilter := make(map[int]bool)
				for _, component := range components {
					for _, idx := range segment.Index.ByComponent[component] {
						if len(matchedIndex) == 0 || matchedIndex[idx] {
							componentFilter[idx] = true
						}
					}
				}
				matchedIndex = componentFilter
			}

			// by host
			if len(hosts) > 0 {
				totalFilters++
				hostFilter := make(map[int]bool)
				for _, host := range hosts {
					for _, idx := range segment.Index.ByHost[host] {
						if len(matchedIndex) == 0 || matchedIndex[idx] {
							hostFilter[idx] = true
						}
					}
				}
				matchedIndex = hostFilter
			}

			// by reqID
			if len(reqIDs) > 0 {
				totalFilters++
				reqFilter := make(map[int]bool)
				for _, reqID := range reqIDs {
					for _, idx := range segment.Index.ByReqID[reqID] {
						if len(matchedIndex) == 0 || matchedIndex[idx] {
							reqFilter[idx] = true
						}
					}
				}
				matchedIndex = reqFilter
			}

			var localResult []model.LogEntry

			// No filters only time range
			if totalFilters == 0 {
				for _, entry := range segment.LogEntries {
					if !startTime.IsZero() && entry.Time.Before(startTime) {
						continue
					}
					if !endTime.IsZero() && entry.Time.After(endTime) {
						continue
					}
					localResult = append(localResult, entry)
				}
			} else {
				// Apply time range to matched indexes
				for idx := range matchedIndex {
					entry := segment.LogEntries[idx]
					if !startTime.IsZero() && entry.Time.Before(startTime) {
						continue
					}
					if !endTime.IsZero() && entry.Time.After(endTime) {
						continue
					}
					localResult = append(localResult, entry)
				}
			}

			mu.Lock()
			result = append(result, localResult...)
			mu.Unlock()

		}(segment)
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Println("Filtering duration:", elapsed)
	return result
}
