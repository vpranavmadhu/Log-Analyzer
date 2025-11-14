package filter

import (
	"os"
	"parser/model"
	"parser/pkg/segment"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func createTestFile(t *testing.T, dir string, name string, entry []string) (string, error) {
	t.Helper()
	path := filepath.Join(dir, name)
	content := strings.Join(entry, "\n")
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp log file: %v\n", err)
	}
	return dir, nil
}

func TestFilterSegmentsWithFilter(t *testing.T) {
	tempDir := t.TempDir()
	name := "file1"
	entry := []string{`2025-10-23 16:53:00.033 | WARN | auth | host=worker01 | request_id=req-nkc44l-7700 | msg="Connection to cache server established"`,
		`2025-10-23 17:03:54.235 | DEBUG | worker | host=web01 | request_id=req-c5xybx-993 | msg="Job completed: send-email"`,
		`2025-10-23 17:03:52.202 | INFO | worker | host=web01 | request_id=req-68rw7u-4614 | msg="Transaction rolled back"`,
		`2025-10-23 17:04:14.387 | DEBUG | api-server | host=worker01 | request_id=req-pbugs0-7474 | msg="Connection pool exhausted"
`}
	path, _ := createTestFile(t, tempDir, name, entry)

	segments, _ := segment.CreateSegments(path)

	got := FilterEntries(segments, []string{"INFO"}, []string{"worker"}, []string{"web01"}, []string{"req-68rw7u-4614"}, time.Time{}, time.Time{})
	if len(got) != 1 {
		t.Errorf("Expected to have one entry after filter, but got: %v", len(got))
	}

	if got[0].Level != "INFO" {
		t.Errorf("Expected level : 'INFO', but got:%v", got[0].Level)
	}

	if got[0].Component != "worker" {
		t.Errorf("Expected level : 'worker', but got:%v", got[0].Component)
	}

}

func TestFilterSegmentsNoFilter(t *testing.T) {
	tempDir := t.TempDir()
	name := "file1"
	entry := []string{`2025-10-23 16:53:00.033 | WARN | auth | host=worker01 | request_id=req-nkc44l-7700 | msg="Connection to cache server established"`,
		`2025-10-23 17:03:54.235 | DEBUG | worker | host=web01 | request_id=req-c5xybx-993 | msg="Job completed: send-email"`,
		`2025-10-23 17:03:52.202 | INFO | worker | host=web01 | request_id=req-68rw7u-4614 | msg="Transaction rolled back"`,
		`2025-10-23 17:04:14.387 | DEBUG | api-server | host=worker01 | request_id=req-pbugs0-7474 | msg="Connection pool exhausted"
`}
	path, _ := createTestFile(t, tempDir, name, entry)

	segments, _ := segment.CreateSegments(path)

	got := FilterEntries(segments, []string{}, []string{}, []string{}, []string{}, time.Time{}, time.Time{})
	if len(got) != 4 {
		t.Errorf("Expected to have one entry after filter, but got: %v", len(got))
	}

	if got[0].Level != "WARN" && got[3].Level != "DEBUG" {
		t.Errorf("Expected level : 'WARN', but got:%v", got[0].Level)
	}

	if got[0].Component != "auth" {
		t.Errorf("Expected level : 'worker', but got:%v", got[0].Component)
	}

}

func TestFilterSegmentsByTime(t *testing.T) {
	tempDir := t.TempDir()
	name := "file1"
	entry := []string{`2025-10-23 16:53:00.033 | WARN | auth | host=worker01 | request_id=req-nkc44l-7700 | msg="Connection to cache server established"`,
		`2025-10-23 17:03:52.202 | INFO | worker | host=web01 | request_id=req-68rw7u-4614 | msg="Transaction rolled back"`,
		`2025-10-23 17:03:54.235 | DEBUG | worker | host=web01 | request_id=req-c5xybx-993 | msg="Job completed: send-email"`,
		`2025-10-23 17:04:14.387 | DEBUG | api-server | host=worker01 | request_id=req-pbugs0-7474 | msg="Connection pool exhausted"
`}
	path, _ := createTestFile(t, tempDir, name, entry)

	segments, _ := segment.CreateSegments(path)
	startTime, _ := time.Parse("2006-01-02 15:04:05", "2025-10-23 17:03:54.235")
	endTime, _ := time.Parse("2006-01-02 15:04:05", "2025-10-23 17:04:14.387")

	got := FilterEntries(segments, []string{}, []string{}, []string{}, []string{}, startTime, endTime)

	if len(got) != 2 {
		t.Errorf("Expected 2 logs between the time range, but got:%d", len(got))
	}

}

func TestEmptySegments(t *testing.T) {
	segments := []model.Segment{}

	got := FilterEntries(segments, nil, nil, nil, nil, time.Time{}, time.Time{})

	if len(got) != 0 {
		t.Errorf("Expected 0 logs for empty segments")
	}

}
