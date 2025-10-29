package parser

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	model "parser/model"
	"regexp"
	"time"
)

func ParseLog(s string) (*model.LogEntry, error) {

	pattern := `^(?P<time>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+)\s+\|\s+(?P<level>[A-Z]+)\s+\|\s+(?P<component>[\w-]+)\s+\|\s+host=(?P<host>[\w-]+)\s+\|\s+request_id=(?P<request_id>[\w-]+)\s+\|\s+msg="(?P<msg>.*)"$`

	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("Error:%v", err)
	}
	matches := r.FindStringSubmatch(s)

	time, err := time.Parse("2006-01-02 15:04:05.000", matches[r.SubexpIndex("time")])
	if err != nil {
		return nil, fmt.Errorf("Error:%v", err)
	}

	return &model.LogEntry{
		Log:        matches[0],
		Time:       time,
		Level:      model.LogLevel(matches[r.SubexpIndex("level")]),
		Component:  matches[r.SubexpIndex("component")],
		Host:       matches[r.SubexpIndex("host")],
		Request_id: matches[r.SubexpIndex("request_id")],
		Message:    matches[r.SubexpIndex("msg")],
	}, nil
}

func ParseLogFile(path string) ([]model.LogEntry, error) {
	var entries []model.LogEntry

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		entry, err := ParseLog(line)
		if err != nil {
			slog.Error("Error while parsing : ", "error", err)
			continue
		}

		entries = append(entries, *entry)
	}
	return entries, nil
}
