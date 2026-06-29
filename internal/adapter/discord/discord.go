// internal/adapter/discord/discord.go
package discord

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/yuying/intake-agent/internal/adapter"
)

type discordAdapter struct {
	token   string
	session *discordgo.Session
}

func New(token string) adapter.Adapter {
	return &discordAdapter{token: token}
}

func (d *discordAdapter) Name() string { return "discord" }

func (d *discordAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
	dg, err := discordgo.New("Bot " + d.token)
	if err != nil {
		return err
	}
	d.session = dg
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author == nil || m.Author.Bot {
			return
		}
		select {
		case out <- adapter.Message{
			ID:        m.ID,
			Source:    "discord",
			ChannelID: m.ChannelID,
			UserID:    m.Author.ID,
			Text:      m.Content,
			Timestamp: time.Now(),
		}:
		default:
			log.Printf("discord: out channel full, dropping message from %s", m.Author.ID)
		}
	})
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages
	if err := dg.Open(); err != nil {
		return err
	}
	<-ctx.Done()
	return dg.Close()
}

func (d *discordAdapter) Reply(ctx context.Context, msg adapter.Message, text string) error {
	if d.session == nil {
		return errors.New("discord session not started")
	}
	_, err := d.session.ChannelMessageSend(msg.ChannelID, text)
	if err != nil {
		log.Printf("discord reply error: %v", err)
	}
	return err
}
