package adapter

import (
	"context"
	"time"
)

type Message struct {
	ID        string
	Source    string
	ChannelID string
	UserID    string
	Text      string
	Timestamp time.Time
}

type Adapter interface {
	Name() string
	Start(ctx context.Context, out chan<- Message) error
	Reply(ctx context.Context, msg Message, text string) error
}
