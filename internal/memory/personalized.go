package tgptmemory

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/tmc/langchaingo/schema"

	"tgpt/internal/models"
	pkgContext "tgpt/pkg/context"
	pkgErrors "tgpt/pkg/errors"
)

type PersonalizedMemory struct {
	constructor func() schema.Memory
	m           map[models.UserID]schema.Memory
	mu          sync.RWMutex
}

// GetMemoryKey getter for memory key.
func (pe *PersonalizedMemory) GetMemoryKey(ctx context.Context) string {
	mem, err := pe.getPersonalMemory(ctx)
	if err != nil {
		slog.Error("get personal memory", slog.String("error", err.Error()))
		return ""
	}
	return mem.GetMemoryKey(ctx)
}

// MemoryVariables Input keys this memory class will load dynamically.
func (pe *PersonalizedMemory) MemoryVariables(ctx context.Context) []string {
	mem, err := pe.getPersonalMemory(ctx)
	if err != nil {
		slog.Error("get personal memory", slog.String("error", err.Error()))
		return nil
	}
	return mem.MemoryVariables(ctx)
}

// LoadMemoryVariables Return key-value pairs given the text input to the chain.
// If None, return all memories
func (pe *PersonalizedMemory) LoadMemoryVariables(
	ctx context.Context,
	inputs map[string]any,
) (map[string]any, error) {
	mem, err := pe.getPersonalMemory(ctx)
	if err != nil {
		return nil, err
	}
	return mem.LoadMemoryVariables(ctx, inputs)
}

// SaveContext Save the context of this model run to memory.
func (pe *PersonalizedMemory) SaveContext(
	ctx context.Context,
	inputs map[string]any,
	outputs map[string]any,
) error {
	mem, err := pe.getPersonalMemory(ctx)
	if err != nil {
		return err
	}
	return mem.SaveContext(ctx, inputs, outputs)
}

// Clear memory contents.
func (pe *PersonalizedMemory) Clear(ctx context.Context) error {
	mem, err := pe.getPersonalMemory(ctx)
	if err != nil {
		return fmt.Errorf("get personal memory: %w", err)
	}
	err = mem.Clear(ctx)

	if err != nil {
		return fmt.Errorf("clear: %w", err)
	}
	return nil
}

func (pe *PersonalizedMemory) getPersonalMemory(ctx context.Context) (schema.Memory, error) {
	userID, ok := pkgContext.UserIDFromCtx(ctx)
	if !ok {
		return nil, fmt.Errorf("user_id from context: %w", pkgErrors.ErrNotFound)
	}

	return pe.getMemory(userID), nil
}

func (pe *PersonalizedMemory) getMemory(userID models.UserID) schema.Memory {
	pe.mu.RLock()
	m, ok := pe.m[userID]
	pe.mu.RUnlock()
	if ok {
		return m
	}

	m = pe.constructor()

	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.m[userID] = m

	return m
}

func NewPersonalized(constructor func() schema.Memory) *PersonalizedMemory {
	return &PersonalizedMemory{
		constructor: constructor,
		m:           map[models.UserID]schema.Memory{},
		mu:          sync.RWMutex{},
	}
}
