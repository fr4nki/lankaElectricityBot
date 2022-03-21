package bot

import (
	"fmt"
	"github.com/fr4nki/lnkElectricityBot/bot/config"
	"github.com/fr4nki/lnkElectricityBot/bot/fetcher"
	"github.com/fr4nki/lnkElectricityBot/bot/helpers"
	"github.com/robfig/cron/v3"
	"gopkg.in/telebot.v3"
	"log"
	"regexp"
	"time"
)

const (
	botHandlerSchedule = "/getschedule"
	botHandleStart     = "/start"
)

func Init() {
	tz, tzError := helpers.GetCurrentTimeZone()
	if tzError != nil {
		log.Fatalf("Cannot parse timezone")
	}

	c := cron.New(cron.WithLocation(tz))

	cfg, cfgErr := config.Get()
	if cfgErr != nil {
		log.Fatalf("Config parsing error: \"%v\"", cfgErr)
	}

	settings := telebot.Settings{
		Token:  cfg.Token,
		Poller: &telebot.LongPoller{Timeout: 15 * time.Second},
	}

	bot, botErr := telebot.NewBot(settings)
	if botErr != nil {
		log.Fatalf("Unable to initiate connection with Telegram API. Error: \"%v\"", botErr)
	}

	bot.Handle(botHandlerSchedule, func(ctx telebot.Context) error {
		area := ctx.Message().Payload

		matched, err := regexp.MatchString(`^[aA-zZ]$`, area)
		if err != nil || matched == false {
			return ctx.Send("Ошибка: название группы указано неверно")
		}

		res, resErr := fetcher.Forecast(area)
		if resErr != nil {
			return ctx.Send("Ошибка: не удалось получить список отключений электричества")
		}

		txt := fmt.Sprintf("%v", res)
		return ctx.Send(txt)
	})

	bot.Handle(botHandleStart, func(ctx telebot.Context) error {
		groupDoc := fmt.Sprintf("[здесь](%s)", "https://drive.google.com/file/d/1L89wIf27TQkqwGrtjKL6lomLDdiFhYz5/view")
		feedback := fmt.Sprintf("[форму](%s)", "https://forms.gle/z8QdPd9yBXLF462WA")

		txt := fmt.Sprintf(`
Напиши боту *%s* с указанием своей группы, чтобы узнать время отключения электричества. Например, "%s L".

Свою группу можно узнать %s.

В случае возникновения проблем, заполните %s. 
		`, botHandlerSchedule, botHandlerSchedule, groupDoc, feedback)

		return ctx.Send(txt, &telebot.SendOptions{ParseMode: "Markdown", DisableWebPagePreview: true})
	})

	c.AddFunc("00 06 * * *", func() {
		for _, chatID := range cfg.ChatsToNotify {
			txt := ""

			fmt.Printf("Sending to %d", chatID)

			tgChatId := telebot.ChatID(chatID)
			res, resErr := fetcher.Forecast(cfg.ElectricityArea)
			if resErr != nil {
				txt += "Ошибка: не удалось получить список отключений электричества"
				bot.Send(tgChatId, txt)
				return
			}

			txt += fmt.Sprintf("%v", res)

			bot.Send(tgChatId, txt)
		}
	})

	c.Start()
	bot.Start()
}
