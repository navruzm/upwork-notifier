package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

var (
	config         Config
	configFileFlag = flag.String("config", "./config.json", "Config file")
	items          = map[string]struct{}{}
)

func main() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(os.Stdout)
	flag.Parse()

	if os.Getenv("CONFIG") != "" {
		*configFileFlag = os.Getenv("CONFIG")
	}

	jsonFile, err := os.Open(*configFileFlag)
	if err != nil {
		log.Panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &config)

	if len(config.Urls) == 0 {
		log.Panic("Must provide url")
	}

	check(true)
	ticker := time.NewTicker(time.Minute * 1)
	for range ticker.C {
		check(false)
	}
}

func check(first bool) {
	log.Info("Checking for available jobs")
	fp := gofeed.NewParser()
	for _, url := range config.Urls {
		feed, err := fp.ParseURL(url)
		if err != nil {
			log.Error("Error parsing url:", err)
			return
		}
		for _, x := range feed.Items {
			if _, ok := items[x.GUID]; !ok && !first {
				if ignore(x.Title + x.Description) {
					continue
				}

				err := beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
				if err != nil {
					log.Error("Error beeping:", err)
					return
				}

				err = beeep.Notify("New job", fmt.Sprintf("%s: %s", x.Title, x.GUID), "information.png")
				if err != nil {
					log.Error("Error notifying:", err)
					return
				}

				log.Infof("%s sended\n", x.Title)
			}
			items[x.GUID] = struct{}{}
		}
	}
}

func ignore(s string) bool {
	for _, i := range config.IgnoredKeywords {
		if strings.Contains(strings.ToLower(s), i) {
			return true
		}
	}
	return false
}

type Config struct {
	IgnoredKeywords []string `json:"ignored_keywords"`
	Urls            []string `json:"urls"`
}
