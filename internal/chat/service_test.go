package chat

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"tgpt/internal/models"
)

func TestChat(t *testing.T) {
	ctx := context.Background()
	userID := models.UserID{ID: "s1kai"}
	s, err := NewService(Config{
		ModelName:  ModelLlama3,
		OllamaAddr: "http://localhost:11434",
		KeepAlive:  "5m",
		QdrantAddr: "http://localhost:6333",
		HTTPClient: http.DefaultClient,
	})
	require.NoError(t, err)
	err = s.HandleQuery(
		ctx,
		models.Message{
			Text:     "this autumn i've been in dubai and it was amazing",
			UserName: userID,
			TimeSend: time.Now(),
			Topic:    "travel",
		},
		emptyHandler,
	)
	require.NoError(t, err)

	err = s.HandleQuery(
		ctx,
		models.Message{
			Text:     "two days ago i back from vladiostok and that trip was horrible",
			UserName: userID,
			TimeSend: time.Now(),
			Topic:    "travel",
		},
		emptyHandler,
	)
	require.NoError(t, err)
	err = s.recall(ctx, models.Message{
		TimeSend:     time.Now(),
		UserName:     userID,
		FromUserName: userID,
		Text:         "summurize all my travels this year",
		Topic:        "travel",
		Command:      commandPrefixRu_RU,
	}, func(ctx context.Context, chunk []byte) error {
		t.Log(string(chunk))
		return nil
	})

	require.NoError(t, err)
}