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
	QRCode      string `json:"qrCode"`
}

type linkOverview struct {
	Code           string    `json:"code"`
	OriginalURL    string    `json:"originalUrl"`
	CreatedAt      time.Time `json:"createdAt"`
	TotalClicks    int       `json:"totalClicks"`
	UniqueVisitors int       `json:"uniqueVisitors"`
}

type linkDetailsResponse struct {
	Code           string         `json:"code"`
	ShortURL       string         `json:"shortUrl"`
	OriginalURL    string         `json:"originalUrl"`
	CreatedAt      time.Time      `json:"createdAt"`
	TotalClicks    int            `json:"totalClicks"`
	UniqueVisitors int            `json:"uniqueVisitors"`
	LastAccessed   *time.Time     `json:"lastAccessed,omitempty"`
	CountryCounts  map[string]int `json:"countryCounts"`
	QRCode         string         `json:"qrCode"`
}
