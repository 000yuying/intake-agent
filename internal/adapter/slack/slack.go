// internal/adapter/slack/slack.go
package slack

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/yuying/intake-agent/internal/adapter"
	"github.com/yuying/intake-agent/internal/engine"
)

type slackAdapter struct {
	signingSecret string
	botToken      string
	engine        *engine.ConfirmEngine
	mux           *http.ServeMux
	out           chan<- adapter.Message
}

// New creates a Slack adapter that registers its HTTP handler on the default ServeMux.
func New(signingSecret, botToken string, eng *engine.ConfirmEngine) adapter.Adapter {
	return NewWithMux(signingSecret, botToken, eng, http.DefaultServeMux)
}

// NewWithMux creates a Slack adapter that registers its HTTP handler on the given ServeMux.
// This is useful for testing without polluting the default mux.
func NewWithMux(signingSecret, botToken string, eng *engine.ConfirmEngine, mux *http.ServeMux) adapter.Adapter {
	return &slackAdapter{
		signingSecret: signingSecret,
		botToken:      botToken,
		engine:        eng,
		mux:           mux,
	}
}

func (s *slackAdapter) Name() string { return "slack" }

func (s *slackAdapter) Start(ctx context.Context, out chan<- adapter.Message) error {
	s.out = out
	s.mux.HandleFunc("/slack/events", s.handleEvent)
	return nil
}

func (s *slackAdapter) handleEvent(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sv, err := slack.NewSecretsVerifier(r.Header, s.signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if _, err = sv.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err = sv.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var cr slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &cr); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cr.Challenge))
		return

	case slackevents.CallbackEvent:
		innerEvent := eventsAPIEvent.InnerEvent
		if msg, ok := innerEvent.Data.(*slackevents.MessageEvent); ok {
			if s.out != nil {
				s.out <- adapter.Message{
					ID:        msg.TimeStamp,
					Source:    "slack",
					ChannelID: msg.Channel,
					UserID:    msg.User,
					Text:      msg.Text,
					Timestamp: time.Now(),
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (s *slackAdapter) Reply(ctx context.Context, msg adapter.Message, text string) error {
	api := slack.New(s.botToken)
	_, _, err := api.PostMessage(msg.ChannelID, slack.MsgOptionText(text, false))
	if err != nil {
		log.Printf("slack reply error: %v", err)
	}
	return err
}
