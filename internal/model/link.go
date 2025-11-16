package model

import "time"

type Link struct {
	Code        string              `json:"code"`
	OriginalURL string              `json:"originalUrl"`
	CreatedAt   time.Time           `json:"createdAt"`
	Clicks      []Click             `json:"-"`
	UniqueIPs   map[string]struct{} `json:"-"`
}

type Click struct {
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	Country   string    `json:"country"`
	UserAgent string    `json:"userAgent"`
}
