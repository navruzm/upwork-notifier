package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

var (
	config         Config
	bot            *tgbotapi.BotAPI
	configFileFlag = flag.String("config", "./config.json", "Config file")
	items          = map[string]interface{}{}
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

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &config)

	if len(config.Urls) == 0 {
		log.Panic("Must provide url")
	}

	if config.Token == "" {
		log.Panic("Must provide key")
	}
	bot, err = tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}
	// bot.Debug = true
	log.Infof("Authorized on account %s", bot.Self.UserName)

	if config.ChatID == 0 {
		log.Panic("Must provide Telegram chat id")
	}
	check(true)
	ticker := time.NewTicker(time.Minute * 1)
	for range ticker.C {
		check(false)
	}
}

func check(first bool) {
	fp := gofeed.NewParser()
	for _, url := range config.Urls {
		feed, _ := fp.ParseURL(url)
		for _, x := range feed.Items {
			if _, ok := items[x.GUID]; !ok && !first {
				if ignore(x.Title + x.Description) {
					continue
				}
				m := tgbotapi.NewMessage(config.ChatID, fmt.Sprintf("[%s](%s)", x.Title, x.GUID))
				m.ParseMode = "Markdown"
				_, err := bot.Send(m)
				if err != nil {
					log.Error(err)
					continue
				}
				log.Info("%s sended\n", x.Title)
			}
			items[x.GUID] = true
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
	ChatID          int64    `json:"chat_id"`
	Token           string   `json:"token"`
	IgnoredKeywords []string `json:"ignored_keywords"`
	Urls            []string `json:"urls"`
}
