package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

var bot *tgbotapi.BotAPI
var chatID int64
var urls []string
var items = map[string]interface{}{}

func main() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(os.Stdout)

	urlsString := os.Getenv("URLS")
	if urlsString == "" {
		log.Panic("Must provide url")
	}
	for _, u := range strings.Split(urlsString, ",") {
		urls = append(urls, u)
	}

	if len(urls) == 0 {
		log.Panic("Must provide url")
	}

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		log.Panic("Must provide key")
	}
	var err error
	bot, err = tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Panic(err)
	}
	// bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	chatID, err = strconv.ParseInt(os.Getenv("TELEGRAM_CHATID"), 10, 64)
	if err != nil {
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
	for _, url := range urls {
		feed, _ := fp.ParseURL(url)
		for _, x := range feed.Items {
			if _, ok := items[x.GUID]; !ok && !first {
				m := tgbotapi.NewMessage(chatID, fmt.Sprintf("[%s](%s)", x.Title, x.GUID))
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
