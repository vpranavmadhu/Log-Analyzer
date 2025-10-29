package parser

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func createTestFile(entry []string) (string, error) {
	f, err := os.CreateTemp("/tmp", "file1")
	if err != nil {
		fmt.Println("Couldn't create temp file.")
	}
	data := strings.Join(entry, "\n")
	_, err = f.Write([]byte(data))
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func TestParseLogEntry(t *testing.T) {
	entry := `2025-10-23 16:53:00.033 | WARN | auth | host=worker01 | request_id=req-nkc44l-7700 | msg="Connection to cache server established"`

	got, _ := ParseLog(entry)
	expectedTime, _ := time.Parse("2006-01-02 15:04:05.000", "2025-10-23 16:53:00.033")

	if got.Time != expectedTime {
		t.Errorf("Expected: %v, but got: %v", expectedTime, got.Time)
	}

	if got.Level != "WARN" {
		t.Errorf("Expected: 'WARN' but got: '%s'", got.Level)
	}

	if got.Component != "auth" {
		t.Errorf("Expected: 'auth' but got: '%s'", got.Component)
	}

	if got.Host != "worker01" {
		t.Errorf("Expected: 'worker01' but got: '%s'", got.Host)

	}

	if got.Request_id != "req-nkc44l-7700" {
		t.Errorf("Expected: 'eq-nkc44l-7700' but got: '%s'", got.Request_id)
	}

	if got.Message != "Connection to cache server established" {
		t.Errorf("Expected: 'Connection to cache server established' but got: '%s'", got.Message)

	}

	if got.Log != entry {
		t.Errorf("Expected: %v, but got: '%v'", entry, got.Log)
	}

}

func TestParseLogFile(t *testing.T) {
	entry := []string{`2025-10-23 16:53:00.033 | WARN | auth | host=worker01 | request_id=req-nkc44l-7700 | msg="Connection to cache server established"`, `2025-10-23 17:03:54.235 | DEBUG | worker | host=web01 | request_id=req-c5xybx-993 | msg="Job completed: send-email"
`}

	path, _ := createTestFile(entry)

	got, err := ParseLogFile(path)
	expectedTime, _ := time.Parse("2006-01-02 15:04:05.000", "2025-10-23 16:53:00.033")

	if err != nil {
		t.Errorf("Error should be nil , but got: %v", err)
	}

	if got[0].Time != expectedTime {
		t.Errorf("Expected: %v, but got: %v", expectedTime, got[0].Time)
	}

	if got[0].Level != "WARN" {
		t.Errorf("Expected: 'WARN' but got: '%s'", got[0].Level)
	}

	if got[0].Component != "auth" {
		t.Errorf("Expected: 'auth' but got: '%s'", got[0].Component)
	}

	if got[0].Host != "worker01" {
		t.Errorf("Expected: 'worker01' but got: '%s'", got[0].Host)

	}

	if got[0].Request_id != "req-nkc44l-7700" {
		t.Errorf("Expected: 'eq-nkc44l-7700' but got: '%s'", got[0].Request_id)
	}

	if got[0].Message != "Connection to cache server established" {
		t.Errorf("Expected: 'Connection to cache server established' but got: '%s'", got[0].Message)

	}

	if got[0].Log != entry[0] {
		t.Errorf("Expected: %v, but got: '%v'", entry, got[0].Log)
	}

}
