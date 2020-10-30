package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/JamesClonk/feed-notifier/config"
	"github.com/JamesClonk/feed-notifier/log"
	"github.com/mmcdole/gofeed"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// watch feeds, call webhooks with payload
	feedWatcher()
}

func feedWatcher() {
	parser := gofeed.NewParser()
	for {
		for _, notification := range config.Get().Notifications {
			if len(notification.Name) > 0 {
				log.Debugf("going through [%s] feeds ...", notification.Name)
			}

			for _, feed := range notification.Feeds {
				log.Debugf("getting feed items for feed [%v] ...", feed.URL)
				f, _ := parser.ParseURL(feed.URL)

				for _, item := range f.Items {
					// check for item to be newer than lastUpdate and younger than maxAge
					if item.UpdatedParsed.After(feed.LastUpdate) && time.Since(*item.UpdatedParsed) < config.Get().MaxAge {
						title := feed.Name
						if len(title) == 0 {
							title = f.Title
						}
						title = fmt.Sprintf("%s - %s", title, item.Title)
						log.Infof("notify about feed item [%s]", title)
						// TODO: notify here
					}
				}
				feed.LastUpdate = *f.UpdatedParsed
			}
		}

		sleepDuration := time.Duration(config.Get().MaxAge.Minutes()/float64(rand.Intn(5)+5)) * time.Minute
		log.Debugf("sleep for [%v]", sleepDuration)
		time.Sleep(sleepDuration)
	}
}
