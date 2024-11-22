package telegram

import (
	"encoding/json"
	"slices"
	"strings"
	"time"

	"tgpt/internal/models"
)

type APIResponse struct {
	Ok          bool            `json:"ok"`
	Result      json.RawMessage `json:"result,omitempty"`
	ErrorCode   int             `json:"error_code,omitempty"`
	Description string          `json:"description,omitempty"`
}

type Message struct {
	Date int `json:"date"`
	Chat struct {
		LastName  string `json:"last_name"`
		ID        string `json:"id"`
		Type      string `json:"type"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
	} `json:"chat"`
	MessageID string `json:"message_id"`
	From      struct {
		LastName  string `json:"last_name"`
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
	} `json:"from"`
	Text string `json:"text"`
}

type MessageReq struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

func (m MessageReq) toBuisnessModel() models.Message {
	split := strings.Split(m.Message.Text, " ")
	var (
		command string
		topic   string

		commandIdx int
		topicIdx   int
	)
	for _, word := range split {
		if strings.HasPrefix(word, "!") {
			command = word
		}

		if strings.HasPrefix(word, "#") {
			topic = word
		}
	}
	split = slices.Delete(split, commandIdx, commandIdx+1)
	split = slices.Delete(split, topicIdx, topicIdx+1)
	text := strings.Join(split, " ")

	if topic == "" {
		topic = "#default"
	}

	return models.Message{
		TimeSend:     time.Now(),
		UserName:     models.UserID{ID: models.ID(m.Message.Chat.Username)},
		FromUserName: models.UserID{ID: models.ID(m.Message.From.Username)},
		Text:         text,
		Topic:        topic,
		Command:      command,
	}
}
