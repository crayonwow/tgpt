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
		port             = os.Getenv("PORT")
		token            = os.Getenv("TOKEN")
		chatID           = os.Getenv("CHAT_ID")
		secretToken      = os.Getenv("SECRET_TOKEN")
		userWhiteListRaw = os.Getenv("USER_WHITE_LIST")
		ollamaAddr       = os.Getenv("OLLAMA_ADDR")
		qdrantAddr       = os.Getenv("QDRANT_ADDR")
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
	if ollamaAddr == "" {
		slog.Error("OLLAMA_ADDR is empty")
		os.Exit(1)
	}
	if qdrantAddr == "" {
		slog.Error("QDRANT_ADDR is empty")
		os.Exit(1)
	}

	httpClient := pkgHttp.NewHttpClient()

	c, err := chat.NewService(chat.Config{
		HTTPClient: httpClient,
		ModelName:  chat.ModelLlama2uncensored,
		OllamaAddr: ollamaAddr,
		KeepAlive:  "1m",
		QdrantAddr: qdrantAddr,
	})
	if err != nil {
		slog.Error("failed to create chat service", "error", err)
		os.Exit(1)
	}
	b := telegram.NewBot(httpClient, token)
	h := telegram.NewHandler(c, b, token, strings.Split(userWhiteListRaw, ","))

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
