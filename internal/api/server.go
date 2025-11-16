package api

import (
	"net/http"
	"strings"

	"link-shortener/internal/storage"
)

const (
	minCodeLength = 6
	maxCodeLength = 8
)

type Config struct {
	Store   storage.Store
	BaseURL string
}

type Server struct {
	store   storage.Store
	baseURL string
}

func NewServer(cfg Config) *Server {
	return &Server{
		store:   cfg.Store,
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/shorten", s.handleShorten)
	mux.HandleFunc("/api/links", s.handleListLinks)
	mux.HandleFunc("/api/links/", s.handleLinkDetails)
	return jsonMiddleware(mux)
}
