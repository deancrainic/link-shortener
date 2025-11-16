package storage

import (
	"errors"

	"link-shortener/internal/model"
)

var (
	ErrCodeExists = errors.New("short code already exists")
	ErrNotFound   = errors.New("link not found")
)

type Store interface {
	Save(link *model.Link) error
	Get(code string) (*model.Link, bool)
	List() []*model.Link
	RecordClick(code string, click model.Click) (*model.Link, error)
}
