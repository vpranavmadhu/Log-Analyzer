package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"parser/model"
	"regexp"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type queryComponent struct {
	key      string
	value    []string
	operator string
}

type Entry struct {
	gorm.Model
	TimeStamp time.Time
	Level     string
	Component string
	Host      string
	RequestId string
	Message   string
}

func (l Entry) String() string {
	if l.TimeStamp.IsZero() {
		return "Empty"
	} else {
		return fmt.Sprintf("%s | %s | %s | %s | %s | %s", l.TimeStamp, l.Level, l.Component, l.Host, l.RequestId, l.Message)

	}
}

func CreateDB(dbUrl string) (*gorm.DB, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level (Silent, Error, Warn, Info)
			IgnoreRecordNotFoundError: false,       // Ignore ErrRecordNotFound error
			Colorful:                  true,        // Enable color
		},
	)

	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, fmt.Errorf("couldn't open database %v", err)
	}
	return db, nil
}

func InitDb(db *gorm.DB) error {
	db.AutoMigrate(&Entry{})
	return nil
}

func AddEntry(db *gorm.DB, e model.LogEntry) error {

	x := Entry{
		TimeStamp: e.Time,
		Level:     string(e.Level),
		Component: e.Component,
		Host:      e.Host,
		RequestId: e.Request_id,
		Message:   e.Message,
	}

	ctx := context.Background()
	err := gorm.G[Entry](db).Create(ctx, &x)
	if err != nil {
		return err
	}
	return nil
}

func parseQuery(parts []string) ([]queryComponent, error) {

	var ret []queryComponent
	// parts := strings.Fields(query)

	// pattern := `^(?P<key>[^\s=!<>]+)\s*(?P<operator>=|!=|>=|<=|>|<)\s*(?P<value>[^,\s]+(?:,[^,\s]+)*)$`
	pattern := `^(?P<key>[^\s=!<>]+)\s*(?P<operator>=|!=|>=|<=|>|<)\s*(?P<value>.+)$`

	r, _ := regexp.Compile(pattern)
	for _, part := range parts {
		matches := r.FindStringSubmatch(part)
		if matches == nil {
			return nil, fmt.Errorf("invalid condition: %s", part)
		}

		val := strings.Split(matches[r.SubexpIndex("value")], ",")
		cond := queryComponent{
			key:      matches[r.SubexpIndex("key")],
			operator: matches[r.SubexpIndex("operator")],
			value:    val,
		}
		ret = append(ret, cond)
	}
	return ret, nil

}

func Query(db *gorm.DB, queryList []string) ([]Entry, error) {
	var ret []Entry

	// Parse the query string
	parsed, err := parseQuery(queryList)
	if err != nil {
		return nil, err
	}

	fmt.Println("Parsed conditions:", parsed)

	q := db
	for _, c := range parsed {
		if len(c.value) == 1 {
			// single value
			fmt.Printf("Applying condition: %s %s %s\n", c.key, c.operator, c.value[0])
			q = q.Where(fmt.Sprintf("%s %s ?", c.key, c.operator), c.value[0])
		} else {
			//multiple values and operator is !=
			if c.operator == "!=" {
				fmt.Printf("Applying NOT IN condition: %s NOT IN %v\n", c.key, c.value)
				q = q.Where(fmt.Sprintf("%s NOT IN ?", c.key), c.value)
			} else {
				// multi value and operator is =
				fmt.Printf("Applying IN condition: %s IN %v\n", c.key, c.value)
				q = q.Where(fmt.Sprintf("%s IN ?", c.key), c.value)
			}

		}
	}

	// Execute final query
	if err := q.Find(&ret).Error; err != nil {
		return nil, err
	}

	return ret, nil
}
