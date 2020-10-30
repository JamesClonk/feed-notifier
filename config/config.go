package config

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	config Config
	once   sync.Once
)

type Config struct {
	LogLevel      string        `yaml:"log_level"`
	LogTimestamp  bool          `yaml:"log_timestamp"`
	MaxAge        time.Duration `yaml:"max_age"`
	Notifications []Notification
}

type Notification struct {
	Name     string
	Feeds    []Feed
	Webhooks []Webhook
}

type Feed struct {
	Name       string
	URL        string
	LastUpdate time.Time `yaml:"last_check"`
}

type Webhook struct {
	URL      string
	Template string
}

func Get() *Config {
	once.Do(func() {
		// initialize
		config = Config{
			LogLevel:      "info",
			LogTimestamp:  true,
			Notifications: make([]Notification, 0),
		}

		// load config file
		if _, err := os.Stat("config.yml"); err == nil {
			data, err := ioutil.ReadFile("config.yml")
			if err != nil {
				log.Println("could not load config.yml")
				log.Fatalln(err.Error())
			}
			if err := yaml.Unmarshal(data, &config); err != nil {
				log.Println("could not parse config.yml")
				log.Fatalln(err.Error())
			}
		}
	})
	return &config
}
