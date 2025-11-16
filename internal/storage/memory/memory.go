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
	if link.UniqueIPs == nil {
		link.UniqueIPs = make(map[string]struct{})
	}
	s.links[link.Code] = link
	return nil
}

func (s *Store) Upsert(link *model.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if link.UniqueIPs == nil {
		link.UniqueIPs = make(map[string]struct{})
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

func (s *Store) List() []*model.Link {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]*model.Link, 0, len(s.links))
	for _, link := range s.links {
		items = append(items, link)
	}
	return items
}

func (s *Store) RecordClick(code string, click model.Click) (*model.Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	link, ok := s.links[code]
	if !ok {
		return nil, storage.ErrNotFound
	}

	link.Clicks = append(link.Clicks, click)
	if link.UniqueIPs == nil {
		link.UniqueIPs = make(map[string]struct{})
	}
	if click.IP != "" {
		link.UniqueIPs[click.IP] = struct{}{}
	}
	return link, nil
}
