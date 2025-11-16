package storage

import (
	"errors"

	"link-shortener/internal/model"
)

var (
	ErrCodeExists = errors.New("short code already exists")
)

type Store interface {
	Save(link *model.Link) error
	Get(code string) (*model.Link, bool)
}
