package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

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
	case "partner_time":
		b.fromPartnerTime(msg)
	case "partner_utc":
		b.fromPartnerUTC(msg)
	case "admins_time":
		b.currentTimeOfAdmins(msg)
	case "this_in_admins_time":
		b.MyTimeInTimeOfAdmins(msg)
	case "start":
		b.start(msg)
	}
}

func (b *Bot) start(msg *tgbotapi.Message) {
	b.replyWithText(msg, "I am working")
}

func (b *Bot) MyTimeInTimeOfAdmins(msg *tgbotapi.Message) {
	t, err := parseDateTime(strings.TrimSpace(msg.CommandArguments()))
	if err != nil {
		b.replyError(msg, "Error while parsing your date + time", err)
		b.send(tgbotapi.NewMessage(373512635, fmt.Sprintf("Пользователь попробовал всунуть в дату: %v", strings.TrimSpace(msg.CommandArguments()))))
	}

	chatID := msg.Chat.ID
	chCng := tgbotapi.ChatConfig{ChatID: chatID}

	chatAdministratorsConfig := tgbotapi.ChatAdministratorsConfig{chCng}
	administrators, err := b.bot.GetChatAdministrators(chatAdministratorsConfig)
	if err != nil {
		b.replyError(msg, "Error while getting administrators", err)
	}
	memberIDs := make(map[int64]int64)
	memberUsernames := make(map[int64]string)
	memberOffsets := make(map[int64]int)

	message := ""
	for _, chatAdministrator := range administrators {
		userID := chatAdministrator.User.ID
		memberIDs[userID] = userID
		memberUsernames[userID] = chatAdministrator.User.UserName
		memberOffsets[userID], err = b.s.GetOffsetByUserID(userID)
		if err != nil {
			continue
		}

		durationString := strconv.Itoa(memberOffsets[userID]) + "h"
		duration, err := time.ParseDuration(durationString)
		if err != nil {
			fmt.Println(err)
			return
		}
		memberCurrTime := time.Now().Add(duration)
		message += memberUsernames[userID] + " " + memberCurrTime.Format("02.01.2006 15:04")
	}
	b.replyWithText(msg, message)

}

func (b *Bot) currentTimeOfAdmins(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	chCng := tgbotapi.ChatConfig{ChatID: chatID}

	chatAdministratorsConfig := tgbotapi.ChatAdministratorsConfig{chCng}
	administrators, err := b.bot.GetChatAdministrators(chatAdministratorsConfig)
	if err != nil {
		b.replyError(msg, "Error while getting administrators", err)
	}
	memberIDs := make(map[int64]int64)
	memberUsernames := make(map[int64]string)
	memberOffsets := make(map[int64]int)

	message := ""
	for _, chatAdministrator := range administrators {
		userID := chatAdministrator.User.ID
		memberIDs[userID] = userID
		memberUsernames[userID] = chatAdministrator.User.UserName
		memberOffsets[userID], err = b.s.GetOffsetByUserID(userID)
		if err != nil {
			continue
		}

		durationString := strconv.Itoa(memberOffsets[userID]) + "h"
		duration, err := time.ParseDuration(durationString)
		if err != nil {
			fmt.Println(err)
			return
		}
		memberCurrTime := time.Now().Add(duration)
		message += memberUsernames[userID] + " " + memberCurrTime.Format("02.01.2006 15:04")
	}
	b.replyWithText(msg, message)

}

func (b *Bot) fromMyTime(msg *tgbotapi.Message) {
	t, err := parseDateTime(strings.TrimSpace(msg.CommandArguments()))
	if err != nil {
		b.replyError(msg, "Error while parsing your date + time", err)
		b.send(tgbotapi.NewMessage(373512635, fmt.Sprintf("Пользователь попробовал всунуть в дату: %v", strings.TrimSpace(msg.CommandArguments()))))
	}

	offset := calculateOffset(t)
	err = b.s.AddWithUTCOffset(msg.From.ID, msg.From.UserName, offset)
	if err != nil {
		b.replyError(msg, "Error while saving your offset", err)
		return
	}
	b.replyWithText(msg, "Your offset saved "+strconv.Itoa(offset))
}

func (b *Bot) fromPartnerTime(msg *tgbotapi.Message) {
	parts := strings.SplitN(strings.TrimSpace(msg.CommandArguments()), " ", 2)
	if len(parts) != 2 {
		fmt.Println("Error parsing input string")
		return
	}
	userName := parts[0]
	datestring := parts[1]

	t, err := parseDateTime(datestring)
	if err != nil {
		b.replyError(msg, "Error while parsing your date + time", err)
		b.send(tgbotapi.NewMessage(373512635, fmt.Sprintf("Пользователь попробовал всунуть в дату: %v", strings.TrimSpace(msg.CommandArguments()))))
	}

	err = b.s.AddWithUTCOffsetOnlyUsername(userName, calculateOffset(t))
	if err != nil {
		b.replyError(msg, "Error while saving your offset", err)
		return
	}
	b.replyWithText(msg, userName+"'s offset saved "+strconv.Itoa(calculateOffset(t)))
}

func calculateOffset(t time.Time) int {
	offsetH := int(t.Hour() - time.Now().UTC().Hour())
	//TODO: сделать имплементацию с половинчатыми часами
	return offsetH
}

func parseDateTime(input string) (time.Time, error) {
	//почему-то не работает
	layouts := []string{
		"2006-01-02 15:04",
		"Jan 2, 2006 15:04",
		"02.01.2006 15:04",
		"2.01.2006 15:04",
		"2.01.06 15:04",
		"02.01.2006 15.04",
		"2.01.2006 15.04",
		"2.01.06 15.04",
		"02/01/2006 15.04",
		"2/01/2006 15.04",
		"2/01/06 15.04",
		"02/01/2006 15:04",
		"2/01/2006 15:04",
		"2/01/06 15:04",
		"01/02/2006 15:04",
		"01/02/06 15:04",
		"01/2/06 15:04",
		"15:04 2006-01-02",
		"15:04 02.01.2006",
		"15:04 2.01.2006",
		"15:04 2.01.06",
		"15.04 02.01.2006",
		"15.04 2.01.2006",
		"15.04 2.01.06",
		"15.04 02/01/2006",
		"15.04 2/01/2006",
		"15.04 2/01/06",
		"15:04 02/01/2006",
		"15:04 2/01/2006",
		"15:04 2/01/06",
		"15:04 01/02/2006",
		"15:04 01/02/06",
		"15:04 01/2/06",
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, input)
		if err == nil {
			break
		}
	}
	return t, err
}

func (b *Bot) fromUTC(msg *tgbotapi.Message) {
	offset, err := strconv.Atoi(strings.TrimSpace(msg.CommandArguments()))
	if err != nil {
		b.replyWithText(msg, "Wrong offset, enter number from range -12 to 12")
		return
	}
	err = b.s.AddWithUTCOffset(msg.From.ID, msg.From.UserName, offset)
	if err != nil {
		b.replyError(msg, "Error while saving your offset", err)
		return
	}
	b.replyWithText(msg, "Your offset saved "+strconv.Itoa(offset))
}

func (b *Bot) fromPartnerUTC(msg *tgbotapi.Message) {
	parts := strings.SplitN(strings.TrimSpace(msg.CommandArguments()), " ", 2)
	if len(parts) != 2 {
		fmt.Println("Error parsing input string")
		return
	}
	userName := parts[0]
	UTCoffset := parts[1]

	offset, err := strconv.Atoi(UTCoffset)
	if err != nil {
		b.replyWithText(msg, "Wrong offset, enter number from range -12 to 12")
		return
	}
	err = b.s.AddWithUTCOffsetOnlyUsername(userName, offset)
	if err != nil {
		b.replyError(msg, "Error while saving your offset", err)
		return
	}
	b.replyWithText(msg, userName+"'s offset saved "+strconv.Itoa(offset))
}

func (b *Bot) replyWithText(to *tgbotapi.Message, text string) {
	msg := tgbotapi.NewMessage(to.Chat.ID, text)
	msg.ReplyToMessageID = to.MessageID
	msg.ParseMode = tgbotapi.ModeHTML
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
