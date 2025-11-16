package api

import "time"

type shortenRequest struct {
	URL         string `json:"url"`
	CustomAlias string `json:"customAlias"`
}

type shortenResponse struct {
	Code        string `json:"code"`
	ShortURL    string `json:"shortUrl"`
	OriginalURL string `json:"originalUrl"`
}

type linkOverview struct {
	Code        string    `json:"code"`
	OriginalURL string    `json:"originalUrl"`
	CreatedAt   time.Time `json:"createdAt"`
}

type linkDetailsResponse struct {
	Code        string    `json:"code"`
	ShortURL    string    `json:"shortUrl"`
	OriginalURL string    `json:"originalUrl"`
	CreatedAt   time.Time `json:"createdAt"`
}
