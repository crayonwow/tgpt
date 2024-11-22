package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

const apiUrl = "https://api.telegram.org/bot{{token}}/{{method_name}}"

const (
	methodSendMessage = "sendMessage"
	methodEditMessage = "editMessageText"
)

func newUrl(token string, method string) string {
	s := strings.Replace(apiUrl, "{{token}}", token, 1)
	s = strings.Replace(s, "{{method_name}}", method, 1)
	return s
}

type Bot struct {
	client *http.Client
	token  string
}

func NewBot(cli *http.Client, token string) *Bot {
	return &Bot{
		token:  token,
		client: cli,
	}
}

func (b *Bot) SendMessage(
	ctx context.Context,
	chatID, message string,
) (Message, error) {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	go func() {
		defer w.Close()
		defer m.Close()

		for k, v := range map[string]string{
			"chat_id": chatID,
			"text":    message,
		} {
			if err := m.WriteField(k, v); err != nil {
				w.CloseWithError(err)
				return
			}
		}
	}()

	req, err := http.NewRequest(http.MethodPost, newUrl(b.token, methodSendMessage), r)
	if err != nil {
		return Message{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		return Message{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Ok {
		return Message{}, fmt.Errorf("api error: %s", apiResp.Description)
	}

	msg := Message{}
	err = json.Unmarshal(apiResp.Result, &message)
	if err != nil {
		return Message{}, fmt.Errorf("failed to decode message: %w", err)
	}

	return msg, nil
}

func (b *Bot) UpdateMessage(
	ctx context.Context,
	chatID, messageID, message string,
) (Message, error) {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	go func() {
		defer w.Close()
		defer m.Close()

		for k, v := range map[string]string{
			"chat_id":    chatID,
			"text":       message,
			"message_id": messageID,
		} {
			if err := m.WriteField(k, v); err != nil {
				w.CloseWithError(err)
				return
			}
		}
	}()

	req, err := http.NewRequest(http.MethodPost, newUrl(b.token, methodEditMessage), r)
	if err != nil {
		return Message{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	if err != nil {
		return Message{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Ok {
		return Message{}, fmt.Errorf("api error: %s", apiResp.Description)
	}

	msg := Message{}
	err = json.Unmarshal(apiResp.Result, &message)
	if err != nil {
		return Message{}, fmt.Errorf("failed to decode message: %w", err)
	}

	return msg, nil
}
