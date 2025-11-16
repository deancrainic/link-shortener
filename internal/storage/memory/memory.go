package memory

import (
	"sync"

	"link-shortener/internal/model"
	"link-shortener/internal/storage"
)

type Store struct {
	mu    sync.RWMutex
	links map[string]*model.Link
}

func New() *Store {
	return &Store{
		links: make(map[string]*model.Link),
	}
}

func (s *Store) Save(link *model.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.links[link.Code]; exists {
		return storage.ErrCodeExists
	}
	s.links[link.Code] = link
	return nil
}

func (s *Store) Get(code string) (*model.Link, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	link, ok := s.links[code]
	return link, ok
}
