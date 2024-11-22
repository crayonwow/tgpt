package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"tgpt/internal/chat"
	"tgpt/internal/models"
)

type chatService interface {
	HandleQuery(
		ctx context.Context,
		message models.Message,
		handler chat.Handler,
	) error
}

func NewHandler(
	chatService chatService,
	bot *Bot,
	secretToken string,
	userWhiteList []string,
) *Handler {
	return &Handler{
		chatService:   chatService,
		bot:           bot,
		secretToken:   secretToken,
		userWhiteList: userWhiteList,
	}
}

type Handler struct {
	chatService   chatService
	bot           *Bot
	secretToken   string
	userWhiteList []string
}

func (h *Handler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	if token != h.secretToken {
		slog.Error("invalid token", "token", token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cant read body", http.StatusInternalServerError)
		return
	}
	message := &MessageReq{}
	err = json.Unmarshal(b, message)
	if err != nil {
		slog.Error("cant decode payload",
			"error", err.Error(),
			"payload", string(b),
		)
		http.Error(w, "cant parse body", http.StatusBadRequest)
		return
	}
	if !slices.Contains(h.userWhiteList, message.Message.Chat.Username) {
		slog.Error(
			"user not in white list",
			"username", message.Message.Chat.Username,
			"white_list", h.userWhiteList,
		)
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx := r.Context()
	newMessage, err := h.bot.SendMessage(ctx, message.Message.Chat.ID, "thinking...")
	if err != nil {
		slog.Error("send message",
			"error", err.Error(),
			"payload", string(b),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sb := &strings.Builder{}
	ha := func(ctx context.Context, chunk []byte) error {
		sb.Write(chunk)
		_, err = h.bot.UpdateMessage(
			ctx,
			message.Message.Chat.ID,
			newMessage.MessageID,
			sb.String(),
		)
		if err != nil {
			return fmt.Errorf("update message: %w", err)
		}
		return nil
	}

	err = h.chatService.HandleQuery(
		ctx,
		message.toBuisnessModel(),
		ha,
	)

	if err != nil {
		slog.Error("handle query",
			"error", err.Error(),
			"payload", string(b),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
