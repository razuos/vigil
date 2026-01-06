package telegram

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Bot struct {
	Token  string
	ChatID string

	// Command Handlers
	commands map[string]func(args string) string

	// Polling state
	offset int
	stop   chan struct{}
	wg     sync.WaitGroup
}

func NewBot(token, chatID string) *Bot {
	return &Bot{
		Token:    token,
		ChatID:   chatID,
		commands: make(map[string]func(string) string),
		stop:     make(chan struct{}),
	}
}

func (b *Bot) RegisterCommand(cmd string, handler func(string) string) {
	b.commands[cmd] = handler
}

func (b *Bot) Start() {
	if b.Token == "" {
		return
	}
	b.wg.Add(1)
	go b.poll()
}

func (b *Bot) Stop() {
	if b.Token == "" {
		return
	}
	close(b.stop)
	b.wg.Wait()
}

func (b *Bot) Send(message string) {
	if b.Token == "" || b.ChatID == "" {
		slog.Warn("Telegram not configured", "message", message)
		return
	}

	go func() {
		apiURL := "https://api.telegram.org/bot" + b.Token + "/sendMessage"
		data := url.Values{}
		data.Set("chat_id", b.ChatID)
		data.Set("text", message)

		resp, err := http.PostForm(apiURL, data)
		if err != nil {
			slog.Error("Failed to send Telegram message", "error", err)
			return
		}
		defer resp.Body.Close()
	}()
}

type updateResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	Text string `json:"text"`
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
}

func (b *Bot) poll() {
	defer b.wg.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-b.stop:
			return
		case <-ticker.C:
			b.getUpdates()
		}
	}
}

func (b *Bot) getUpdates() {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=1", b.Token, b.offset)
	resp, err := http.Get(apiURL)
	if err != nil {
		slog.Error("Telegram poll error", "error", err)
		return
	}
	defer resp.Body.Close()

	var uRes updateResponse
	if err := json.NewDecoder(resp.Body).Decode(&uRes); err != nil {
		return
	}

	for _, u := range uRes.Result {
		b.offset = u.UpdateID + 1
		if u.Message != nil {
			b.handleMessage(u.Message)
		}
	}
}

func (b *Bot) handleMessage(msg *Message) {
	txt := strings.TrimSpace(msg.Text)
	if !strings.HasPrefix(txt, "/") {
		return
	}

	parts := strings.SplitN(txt, " ", 2)
	cmd := parts[0]
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	if handler, ok := b.commands[cmd]; ok {
		reply := handler(args)
		if reply != "" {
			b.Send(reply)
		}
	}
}
