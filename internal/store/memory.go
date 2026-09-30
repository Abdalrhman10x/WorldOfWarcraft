package store

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"WorldOfWarcraft/internal/domain"
)

var (
	ErrNotFound      = errors.New("character not found")
	ErrAlreadyExists = errors.New("character already exists")
)

type CharacterStore interface {
	Create(name domain.Name, race domain.Race, class domain.Class) (*domain.Character, error)
	Get(name domain.Name) (*domain.Character, error)
	With(name domain.Name, fn func(*domain.Character) error) error
	List() []*domain.Character
	Delete(name domain.Name) error
}

type Memory struct {
	mu    sync.RWMutex
	chars map[domain.Name]*domain.Character
}

func NewMemory() *Memory {
	return &Memory{chars: make(map[domain.Name]*domain.Character)}
}

func (s *Memory) Create(name domain.Name, race domain.Race, class domain.Class) (*domain.Character, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, err := domain.NewCharacter(name, race, class)
	if err != nil {
		return nil, fmt.Errorf("create character: %w", err)
	}
	if _, exists := s.chars[c.Name]; exists {
		return nil, fmt.Errorf("%q: %w", c.Name, ErrAlreadyExists)
	}
	s.chars[c.Name] = c
	return c, nil
}

func (s *Memory) Get(name domain.Name) (*domain.Character, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.chars[name]
	if !ok {
		return nil, fmt.Errorf("%q: %w", name, ErrNotFound)
	}
	return c, nil
}

// With runs fn against the stored character while holding the write lock.
// This is the only safe way to mutate: Get hands out a live pointer, so two
// concurrent requests mutating it would race.
//
// The callback must NOT call back into the store. sync.RWMutex is not
// reentrant, so a Get inside a With deadlocks permanently. Use the character
// passed to the callback; that is the one it locked.
func (s *Memory) With(name domain.Name, fn func(*domain.Character) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.chars[name]
	if !ok {
		return fmt.Errorf("%q: %w", name, ErrNotFound)
	}
	return fn(c)
}

func (s *Memory) List() []*domain.Character {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*domain.Character, 0, len(s.chars))
	for _, c := range s.chars {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func (s *Memory) Delete(name domain.Name) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.chars[name]; !ok {
		return fmt.Errorf("%q: %w", name, ErrNotFound)
	}
	delete(s.chars, name)
	return nil
}
