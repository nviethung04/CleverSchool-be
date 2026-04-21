package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type DiscordHook struct {
	WebhookURL string
	LevelsList []logrus.Level
}

func NewDiscordHook(webhookURL string, levels ...logrus.Level) *DiscordHook {
	return &DiscordHook{
		WebhookURL: webhookURL,
		LevelsList: levels,
	}
}

func (h *DiscordHook) Levels() []logrus.Level {
	return h.LevelsList
}

func (h *DiscordHook) Fire(entry *logrus.Entry) error {
	// Format nội dung gửi sang Discord
	msg := fmt.Sprintf("**[%s]** %s", entry.Level, entry.Message)
	if len(entry.Data) > 0 {
		msg += fmt.Sprintf("\nData: %+v", entry.Data)
	}

	// JSON payload cho Discord webhook
	payload := []byte(fmt.Sprintf(`{"content": %q}`, msg))

	_, err := http.Post(h.WebhookURL, "application/json", bytes.NewBuffer(payload))
	return err
}

type AsyncDiscordHook struct {
    webhookURL string
    ch         chan *logrus.Entry
}

func NewAsyncDiscordHook(webhookURL string) *AsyncDiscordHook {
    h := &AsyncDiscordHook{
        webhookURL: webhookURL,
        ch:         make(chan *logrus.Entry, 100), // buffer
    }

    // worker goroutine
    go func() {
        for entry := range h.ch {
            h.send(entry)
        }
    }()

    return h
}

func (h *AsyncDiscordHook) Levels() []logrus.Level {
    return []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}

func (h *AsyncDiscordHook) Fire(entry *logrus.Entry) error {
    // đẩy entry vào channel, không block
    select {
    case h.ch <- entry:
    default:
        // nếu channel đầy thì bỏ qua hoặc log local
    }
    return nil
}

func (h *AsyncDiscordHook) send(entry *logrus.Entry) {
    payload := map[string]string{
        "content": fmt.Sprintf("**%s**: %s", entry.Level, entry.Message),
    }
    jsonData, _ := json.Marshal(payload)
    http.Post(h.webhookURL, "application/json", bytes.NewBuffer(jsonData))
}

