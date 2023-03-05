package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aakrasnova/zone-mate/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	s   *service.Service
	bot *tgbotapi.BotAPI
}

func NewBot(s *service.Service, token string) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	bot.Debug = true // TODO before release take from config

	return &Bot{s: s, bot: bot}, nil
}

func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for update := range updates {
		if msg := update.Message; msg != nil {
			b.handleMsg(msg)
		}
	}
}

func (b *Bot) Stop() {
	b.bot.StopReceivingUpdates()
}

func (b *Bot) handleMsg(msg *tgbotapi.Message) {
	defer func() {
		if rec := recover(); rec != nil {
			b.send(tgbotapi.NewMessage(373512635, fmt.Sprintf("Я запаниковал: %v", rec)))
		}
	}()

	switch msg.Command() {
	case "from_my_time":
		b.fromMyTime(msg)
	case "from_utc":
		b.fromUTC(msg)
	case "start":
		b.start(msg)
	}
}

func (b *Bot) start(msg *tgbotapi.Message) {
	b.replyWithText(msg, "I am working")
}

func (b *Bot) fromMyTime(msg *tgbotapi.Message) {

}

func (b *Bot) fromUTC(msg *tgbotapi.Message) {
	offset, err := strconv.Atoi(strings.TrimSpace(msg.CommandArguments()))
	if err != nil {
		b.replyWithText(msg, "Wrong offset, enter number from range -12 to 12")
		return
	}
	err = b.s.AddWithUTCOffset(msg.From.ID, offset)
	if err != nil {
		b.replyError(msg, "Error while saving your offset", err)
		return
	}
	b.replyWithText(msg, "Your offset saved")
}

func (b *Bot) replyWithText(to *tgbotapi.Message, text string) {
	msg := tgbotapi.NewMessage(to.Chat.ID, text)
	msg.ReplyToMessageID = to.MessageID
	msg.ParseMode = tgbotapi.ModeHTML
	// msg.ParseMode = tgbotapi.ModeMarkdownV2
	b.send(msg)
}

func (b *Bot) replyError(to *tgbotapi.Message, text string, err error) {
	msg := tgbotapi.NewMessage(to.Chat.ID, text)
	msg.ReplyToMessageID = to.MessageID
	if err != nil {
		log.Println(err.Error())
	}
	b.send(msg)
}

func (b *Bot) send(msg tgbotapi.MessageConfig) {
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Println("error while sending message: ", err)
	}
}
