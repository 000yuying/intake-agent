// internal/adapter/gchat/gchat.go
package gchat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/yuying/intake-agent/internal/adapter"
)

type gchatAdapter struct {
	webhookURL string
	mux        *http.ServeMux
	out        chan<- adapter.Message
}

func New(webhookURL string) adapter.Adapter {
	return NewWithMux(webhookURL, http.DefaultServeMux)
}

func NewWithMux(webhookURL string, mux *http.ServeMux) adapter.Adapter {
	return &gchatAdapter{webhookURL: webhookURL, mux: mux}
}

func (g *gchatAdapter) Name() string { return "gchat" }

func (g *gchatAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
	g.out = out
	g.mux.HandleFunc("/gchat/events", g.handleEvent)
	return nil
}

type chatEvent struct {
	Type    string `json:"type"`
	Message struct {
		Name   string `json:"name"`
		Text   string `json:"text"`
		Sender struct {
			Name string `json:"name"`
		} `json:"sender"`
	} `json:"message"`
	Space struct {
		Name string `json:"name"`
	} `json:"space"`
}

func (g *gchatAdapter) handleEvent(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var event chatEvent
	if err := json.Unmarshal(body, &event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if event.Type != "MESSAGE" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if g.out != nil {
		select {
		case g.out <- adapter.Message{
			ID:        event.Message.Name,
			Source:    "gchat",
			ChannelID: event.Space.Name,
			UserID:    event.Message.Sender.Name,
			Text:      event.Message.Text,
			Timestamp: time.Now(),
		}:
		default:
			log.Printf("gchat: out channel full, dropping message from %s", event.Message.Sender.Name)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (g *gchatAdapter) Reply(ctx context.Context, msg adapter.Message, text string) error {
	payload, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.webhookURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("gchat reply error: %v", err)
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gchat: reply failed with status %d", resp.StatusCode)
	}
	return nil
}
