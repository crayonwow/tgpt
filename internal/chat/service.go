package chat

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/qdrant"

	tgptmemory "tgpt/internal/memory"
	"tgpt/internal/models"
	pkgContext "tgpt/pkg/context"
)

const (
	commandPrefixEn_US = "!bro "
	commandPrefixRu_RU = "!бро "
)

const (
	ModelLlama2uncensored = "llama2-uncensored"
	ModelLlama3           = "llama3.1"
)

type Handler func(ctx context.Context, chunk []byte) error

type Config struct {
	HTTPClient *http.Client

	ModelName  string
	OllamaAddr string
	KeepAlive  string

	QdrantAddr string
}

type Service struct {
	store vectorstores.VectorStore
	llm   llms.Model

	mem schema.Memory
}

func NewService(cfg Config) (*Service, error) {
	_ollama, err := ollama.New(
		ollama.WithModel(cfg.ModelName),
		ollama.WithKeepAlive(cfg.KeepAlive),
		ollama.WithServerURL(cfg.OllamaAddr),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ollama client: %w", err)
	}

	e, err := embeddings.NewEmbedder(_ollama)
	if err != nil {
		return nil, fmt.Errorf("can't build embeder: %w", err)
	}

	qdrantUrl, err := url.Parse(cfg.QdrantAddr)
	if err != nil {
		return nil, fmt.Errorf("can't parse qdrant url: %w", err)
	}

	q, err := qdrant.New(
		qdrant.WithURL(*qdrantUrl),
		qdrant.WithEmbedder(e),
		qdrant.WithCollectionName("chat"),
	)
	if err != nil {
		return nil, fmt.Errorf("can't connect to qdrant: %w", err)
	}
	mem := tgptmemory.NewPersonalized(func() schema.Memory {
		return memory.NewConversationBuffer()
	})

	return &Service{
		store: q,
		llm:   _ollama,
		mem:   mem,
	}, nil
}

func (s *Service) HandleQuery(
	ctx context.Context,
	message models.Message,
	handler Handler,
) error {
	switch message.Command {
	case commandPrefixEn_US, commandPrefixRu_RU:
		return s.recall(ctx, message, handler)
	default:
		return s.handleMessage(ctx, message)
	}
}

func (s *Service) handleMessage(
	ctx context.Context,
	message models.Message,
) error {
	err := s.saveDocument(ctx, message)
	if err != nil {
		return fmt.Errorf("save document: %w", err)
	}

	err = s.remember(ctx, message)
	if err != nil {
		return fmt.Errorf("save document: %w", err)
	}
	return nil
}

func (s *Service) saveDocument(
	ctx context.Context,
	message models.Message,
) error {
	metaData := map[string]any{
		"user_id":      message.UserName.String(),
		"from_user_id": message.FromUserName.String(),
		"topic":        message.Topic,
	}

	_, err := s.store.AddDocuments(
		pkgContext.CtxWithUserID(ctx, message.UserName),
		[]schema.Document{
			{
				PageContent: "'" + message.FromUserName.String() + "': " + message.Text,
				Metadata:    metaData,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("add documents: %w", err)
	}
	return nil
}

func (s *Service) recall(
	ctx context.Context,
	message models.Message,
	handler Handler,
) error {
	filters := filter{
		Must: []filterEntry{
			{
				Key: "topic",
				Match: filterEntryMatch{
					Value: message.Topic,
				},
			},
			{
				Key: "user_id",
				Match: filterEntryMatch{
					Value: message.UserName.String(),
				},
			},
		},
	}

	conv := chains.NewConversationalRetrievalQAFromLLM(
		s.llm,
		vectorstores.ToRetriever(
			s.store,
			10,
			vectorstores.WithFilters(filters),
		),
		s.mem,
	)

	_, err := chains.Call(
		pkgContext.CtxWithUserID(ctx, message.UserName),
		conv,
		map[string]any{
			"question": message.Text,
		},
		chains.WithStreamingFunc(handler),
	)
	if err != nil {
		return fmt.Errorf("call: %w", err)
	}
	return nil
}

func (s *Service) remember(
	ctx context.Context,
	message models.Message,
) error {
	return s.mem.SaveContext(ctx, map[string]any{"": message.Text}, map[string]any{"": ""})
}

type filter struct {
	Must []filterEntry `json:"must,omitempty"`
}

type filterEntry struct {
	Key   string           `json:"key"`
	Match filterEntryMatch `json:"match"`
}

type filterEntryMatch struct {
	Value string `json:"value"`
}

func emptyHandler(_ context.Context, _ []byte) error {
	return nil
}
