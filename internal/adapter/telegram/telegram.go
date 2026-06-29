// internal/adapter/telegram/telegram.go
package telegram

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yuying/intake-agent/internal/adapter"
)

type telegramAdapter struct {
	token string
	mu    sync.Mutex
	bot   *tgbotapi.BotAPI
}

func New(token string) adapter.Adapter {
	return &telegramAdapter{token: token}
}

func (t *telegramAdapter) Name() string { return "telegram" }

func (t *telegramAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
	bot, err := tgbotapi.NewBotAPI(t.token)
	if err != nil {
		return err
	}
	t.mu.Lock()
	t.bot = bot
	t.mu.Unlock()
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
			if update.Message.From == nil {
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
			select {
			case out <- msg:
			default:
				log.Printf("telegram: out channel full, dropping message from %s", msg.UserID)
			}
		}
	}
}

func (t *telegramAdapter) Reply(ctx context.Context, msg adapter.Message, text string) error {
	t.mu.Lock()
	bot := t.bot
	t.mu.Unlock()
	if bot == nil {
		log.Printf("telegram: bot not started, cannot reply")
		return nil
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
