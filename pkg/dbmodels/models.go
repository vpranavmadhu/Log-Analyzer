package models

import (
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

type LogLevel struct {
	Id    uint   `gorm:"primaryKey"`
	Level string `gorm:"unique;not null"`
}

type LogComponent struct {
	Id        uint   `gorm:"primaryKey"`
	Component string `gorm:"unique;not null"`
}

type LogHost struct {
	Id   uint   `gorm:"primaryKey"`
	Host string `gorm:"unique;not null"`
}

type Entry struct {
	gorm.Model
	TimeStamp   time.Time `gorm:"column:timestamp"`
	LevelId     uint
	ComponentId uint
	HostId      uint
	RequestId   string
	Message     string

	Level     LogLevel     `gorm:"foreignKey:LevelId"`
	Component LogComponent `gorm:"foreignKey:ComponentId"`
	Host      LogHost      `gorm:"foreignKey:HostId"`
}

func (e Entry) String() string {
	if e.TimeStamp.IsZero() {
		return "Empty"
	} else {
		return fmt.Sprintf("%s | %s | %s | %s | %s | %s", e.TimeStamp, e.Level.Level, e.Component.Component, e.Host.Host, e.RequestId, e.Message)

	}
}

// connecting database
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

// table creation
func InitDb(db *gorm.DB) error {
	db.AutoMigrate(&LogLevel{}, &LogComponent{}, &LogHost{}, &Entry{})
	for _, l := range []LogLevel{
		{Level: "INFO"},
		{Level: "WARN"},
		{Level: "ERROR"},
		{Level: "DEBUG"},
	} {
		db.FirstOrCreate(&l, l)
	}

	for _, c := range []LogComponent{
		{Component: "api-server"},
		{Component: "database"},
		{Component: "cache"},
		{Component: "worker"},
		{Component: "auth"},
	} {
		db.FirstOrCreate(&c, c)
	}

	for _, h := range []LogHost{
		{Host: "web01"},
		{Host: "web02"},
		{Host: "cache01"},
		{Host: "worker01"},
		{Host: "db01"},
	} {
		db.FirstOrCreate(&h, h)
	}
	return nil
}

func AddEntry(db *gorm.DB, e model.LogEntry) error {

	entry := Entry{
		TimeStamp: e.Time,
		RequestId: e.Request_id,
		Message:   e.Message,
	}

	// string to id --> level
	var level LogLevel
	if err := db.Where("level = ?", e.Level).First(&level).Error; err != nil {
		return err
	}
	entry.LevelId = level.Id

	//component
	var component LogComponent
	if err := db.Where("component = ?", e.Component).First(&component).Error; err != nil {
		return err
	}
	entry.ComponentId = component.Id

	//host
	var host LogHost
	if err := db.Where("host = ?", e.Host).First(&host).Error; err != nil {
		return err
	}
	entry.HostId = host.Id

	//insert the entry
	if err := db.Create(&entry).Error; err != nil {
		return err
	}
	return nil
}

func parseQuery(parts []string) ([]queryComponent, error) {

	var ret []queryComponent
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

	q := db.Preload("Level").Preload("Component").Preload("Host")

	for _, c := range parsed {

		key := strings.ToLower(c.key)

		// logic for foriegn key columns
		switch key {

		case "level":
			var ids []uint
			for _, v := range c.value {
				var lvl LogLevel
				if err := db.First(&lvl, "level = ?", v).Error; err != nil {
					return nil, fmt.Errorf("unknown level '%s'", v)
				}
				ids = append(ids, lvl.Id)
			}
			c.key = "level_id"
			c.value = toStringSlice(ids)

		case "component":
			var ids []uint
			for _, v := range c.value {
				var comp LogComponent
				if err := db.First(&comp, "component = ?", v).Error; err != nil {
					return nil, fmt.Errorf("unknown component '%s'", v)
				}
				ids = append(ids, comp.Id)
			}
			c.key = "component_id"
			c.value = toStringSlice(ids)

		case "host":
			var ids []uint
			for _, v := range c.value {
				var h LogHost
				if err := db.First(&h, "host = ?", v).Error; err != nil {
					return nil, fmt.Errorf("unknown host '%s'", v)
				}
				ids = append(ids, h.Id)
			}
			c.key = "host_id"
			c.value = toStringSlice(ids)

		}

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

// convert int to string
func toStringSlice(nums []uint) []string {
	s := make([]string, len(nums))
	for i, n := range nums {
		s[i] = fmt.Sprint(n)
	}
	return s
}
