package api

type shortenRequest struct {
	URL         string `json:"url"`
	CustomAlias string `json:"customAlias"`
}

type shortenResponse struct {
	Code        string `json:"code"`
	ShortURL    string `json:"shortUrl"`
	OriginalURL string `json:"originalUrl"`
}
