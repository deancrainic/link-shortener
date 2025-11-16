package model

import "time"

// Link represents a shortened URL entry.
type Link struct {
	Code        string    `json:"code"`
	OriginalURL string    `json:"originalUrl"`
	CreatedAt   time.Time `json:"createdAt"`
}
