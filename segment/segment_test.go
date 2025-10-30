package segment

import (
	"os"
	"parser/model"
	"path/filepath"
	"reflect"
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

func TestCreateSegment(t *testing.T) {
	tempDir := t.TempDir()
	name := "file1"
	entry := []string{`2025-10-23 16:53:00.033 | WARN | auth | host=worker01 | request_id=req-nkc44l-7700 | msg="Connection to cache server established"`,
		`2025-10-23 17:03:54.235 | DEBUG | worker | host=web01 | request_id=req-c5xybx-993 | msg="Job completed: send-email"`,
		`2025-10-23 17:03:52.202 | INFO | worker | host=web01 | request_id=req-68rw7u-4614 | msg="Transaction rolled back"`,
		`2025-10-23 17:04:14.387 | DEBUG | api-server | host=worker01 | request_id=req-pbugs0-7474 | msg="Connection pool exhausted"
`}
	path, _ := createTestFile(t, tempDir, name, entry)

	got, err := CreateSegments(path)
	expectedStartTime, _ := time.Parse("2006-01-02 15:04:05.000", "2025-10-23 16:53:00.033")
	expectedEndTime, _ := time.Parse("2006-01-02 15:04:05.000", "2025-10-23 17:04:14.387")

	expectedSegmentIndex := model.SegmentIndex{
		ByLevel: map[string][]int{
			"WARN":  {0},
			"DEBUG": {1, 3},
			"INFO":  {2},
		},
		ByComponent: map[string][]int{
			"auth":       {0},
			"worker":     {1, 2},
			"api-server": {3},
		},
	}

	if err != nil {
		t.Errorf("Expected Error to be nil, but got: %v", err)
	}

	if got[0].FileName != name {
		t.Errorf("Expected %v to start with file1, but got:%v", name, got[0].FileName)
	}

	if got[0].StartTime != expectedStartTime {
		t.Errorf("Expected start time: %v, but got: %v", expectedStartTime, got[0].StartTime)
	}

	if got[0].EndTime != expectedEndTime {
		t.Errorf("Expected end time: %v, but got: %v", expectedEndTime, got[0].EndTime)
	}

	if reflect.DeepEqual(got[0].Index, expectedSegmentIndex) {
		t.Errorf("Expected: %v, but got: %v", expectedSegmentIndex, got[0].Index)
	}
}

func TestCreateSegmentsWrongDir(t *testing.T) {
	path := "worngpath"
	_, err := CreateSegments(path)

	if err == nil {
		t.Errorf("Expected error:%v, but got no error", err)
	}
}
