package api

import (
	"net/http"
	"strings"
	"time"

	"link-shortener/internal/storage"
)

const (
	minCodeLength     = 6
	maxCodeLength     = 8
	rateLimitRequests = 10
	rateLimitWindow   = time.Minute
)

type Config struct {
	Store   storage.Store
	BaseURL string
}

type Server struct {
	store   storage.Store
	baseURL string
	limiter *rateLimiter
}

func NewServer(cfg Config) *Server {
	return &Server{
		store:   cfg.Store,
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
		limiter: newRateLimiter(rateLimitRequests, rateLimitWindow),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/shorten", s.handleShorten)
	mux.HandleFunc("/api/links", s.handleListLinks)
	mux.HandleFunc("/api/links/", s.handleLinkDetails)
	mux.HandleFunc("/", s.handleRedirect)
	return s.rateLimitMiddleware(jsonMiddleware(mux))
}
