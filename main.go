package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/JamesClonk/feed-notifier/config"
	"github.com/JamesClonk/feed-notifier/log"
	md "github.com/JohannesKaufmann/html-to-markdown"
	md_plugin "github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/Masterminds/sprig"
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

						for _, hook := range notification.Webhooks {
							// try to convert feed item content to markdown
							markdown, err := md.NewConverter("", true, nil).Use(md_plugin.GitHubFlavored()).ConvertString(item.Content)
							if err != nil {
								log.Errorf("could not convert feed item content to markdown: %v", err)
							}

							// parse webhook template, fill it with data
							var data bytes.Buffer
							tmpl := template.Must(template.New("webhook").Funcs(sprig.FuncMap()).Parse(hook.Template))
							if err := tmpl.Execute(
								&data,
								struct {
									Title           string
									Feed            *gofeed.Feed
									Item            *gofeed.Item
									MarkdownContent string
								}{
									Title:           title,
									Feed:            f,
									Item:            item,
									MarkdownContent: markdown,
								},
							); err != nil {
								log.Errorf("could not prepare webhook template for [%s]: %v", hook.URL, err)
								continue
							}

							// post parsed template to webhook URL
							func() {
								resp, err := http.Post(hook.URL, "application/json", &data)
								if err != nil {
									log.Errorf("could not post notification of [%s] to [%s]: %v", title, hook.URL, err)
									return
								}
								defer resp.Body.Close()

								body, err := ioutil.ReadAll(resp.Body)
								if err != nil {
									log.Errorf("could not read post body for notification of [%s] to [%s]: %v", title, hook.URL, err)
									return
								}

								if resp.StatusCode != 200 {
									log.Errorf("notification failed: %s", string(body))
								} else {
									log.Infoln("notification success!")
								}
							}()
						}
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
