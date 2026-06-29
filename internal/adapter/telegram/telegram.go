// internal/adapter/telegram/telegram.go
package telegram

import (
	"context"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yuying/intake-agent/internal/adapter"
	"github.com/yuying/intake-agent/internal/engine"
)

type telegramAdapter struct {
	token  string
	engine *engine.ConfirmEngine
}

func New(token string, eng *engine.ConfirmEngine) adapter.Adapter {
	return &telegramAdapter{token: token, engine: eng}
}

func (t *telegramAdapter) Name() string { return "telegram" }

func (t *telegramAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
	bot, err := tgbotapi.NewBotAPI(t.token)
	if err != nil {
		return err
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			msg := adapter.Message{
				ID:        strconv.Itoa(update.Message.MessageID),
				Source:    "telegram",
				ChannelID: strconv.FormatInt(update.Message.Chat.ID, 10),
				UserID:    strconv.FormatInt(update.Message.From.ID, 10),
				Text:      update.Message.Text,
				Timestamp: time.Unix(int64(update.Message.Date), 0),
			}
			out <- msg
		}
	}
}

func (t *telegramAdapter) Reply(ctx context.Context, msg adapter.Message, text string) error {
	bot, err := tgbotapi.NewBotAPI(t.token)
	if err != nil {
		return err
	}
	chatID, err := strconv.ParseInt(msg.ChannelID, 10, 64)
	if err != nil {
		return err
	}
	reply := tgbotapi.NewMessage(chatID, text)
	_, err = bot.Send(reply)
	if err != nil {
		log.Printf("telegram reply error: %v", err)
	}
	return err
}
