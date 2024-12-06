package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"

	"tgpt/internal/chat"
	"tgpt/internal/telegram"
	pkgHttp "tgpt/pkg/http"
)

func main() {
	var (
		port             = os.Getenv("HTTP_PORT")
		token            = os.Getenv("TELEGRAM_BOT_TOKEN")
		chatID           = os.Getenv("TELEGRAM_CHAT_ID")
		secretToken      = os.Getenv("TELEGRAM_SECRET_TOKEN")
		userWhiteListRaw = os.Getenv("TELEGRAM_USER_WHITE_LIST")
		qdrantAddr       = os.Getenv("QDRANT_ADDR")

		modelType = os.Getenv("MODEL_TYPE")

		ollamaAddr = os.Getenv("OLLAMA_ADDR")
		chatGPTKey = os.Getenv("CHAT_GPT_KEY")
	)

	if token == "" {
		slog.Error("token is empty")
		os.Exit(1)
	}
	if chatID == "" {
		slog.Error("chatID is empty")
		os.Exit(1)
	}
	if port == "" {
		slog.Error("PORT is empty")
		os.Exit(1)
	}

	if secretToken == "" {
		slog.Error("SECRET_TOKEN is empty")
		os.Exit(1)
	}

	if userWhiteListRaw == "" || len(strings.Split(userWhiteListRaw, ",")) == 0 {
		slog.Error("WHITE_LIST is empty")
		os.Exit(1)
	}
	if qdrantAddr == "" {
		slog.Error("QDRANT_ADDR is empty")
		os.Exit(1)
	}

	httpClient := pkgHttp.NewHttpClient()

	c, err := chat.NewService(chat.Config{
		ModelType:  modelType,
		ModelName:  chat.ModelLlama2uncensored,
		OllamaAddr: ollamaAddr,
		KeepAlive:  "1m",
		QdrantAddr: qdrantAddr,
		ChatGPTKey: chatGPTKey,
	})
	if err != nil {
		slog.Error("failed to create chat service", "error", err)
		os.Exit(1)
	}
	b := telegram.NewBot(httpClient, token)
	h := telegram.NewHandler(c, b, secretToken, strings.Split(userWhiteListRaw, ","))

	router := http.NewServeMux()
	router.HandleFunc("/webhook", h.HandleMessage)

	srv := NewServer(port, router)
	err = srv.ListenAndServe()
	if err != nil {
		slog.Error("cant start server", "error", err)
		os.Exit(1)
	}
}

func NewServer(port string, h http.Handler) *http.Server {
	address := net.JoinHostPort("0.0.0.0", port)

	srv := &http.Server{
		Addr:    address,
		Handler: h,
	}

	return srv
}
