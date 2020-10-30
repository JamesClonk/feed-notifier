package main

import (
	"math/rand"
	"time"

	"github.com/JamesClonk/feed-notifier/config"
	"github.com/JamesClonk/feed-notifier/log"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// watch feeds, call webhooks with payload
	feedWatcher()
}

func feedWatcher() {
	for {
		for _, notification := range config.Get().Notifications {
			for _, feed := range notification.Feeds {
				log.Infof("getting feed entries for feed [%v] ...", feed.URL)
				// TODO: get feed entries here
				feedChecked := time.Now()

				feedEntryTimestamp := time.Now()
				if feedEntryTimestamp.After(feed.LastCheck) && time.Since(feedEntryTimestamp) < config.Get().MaxAge {
					// TODO: notify here
					feed.LastCheck = feedChecked
				}
			}
		}

		sleepDuration := time.Duration(config.Get().MaxAge.Minutes()/float64(rand.Intn(5)+5)) * time.Minute
		log.Debugf("sleep for [%v]", sleepDuration)
		time.Sleep(sleepDuration)
	}
}
