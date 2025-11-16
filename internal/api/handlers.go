package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"link-shortener/internal/model"
	"link-shortener/internal/storage"
)

func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	originalURL, err := validateURL(payload.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid url: %v", err), http.StatusBadRequest)
		return
	}

	code, err := s.resolveCode(payload.CustomAlias)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.ErrCodeExists) || errors.Is(err, errInvalidCustomCode) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	link := &model.Link{
		Code:        code,
		OriginalURL: originalURL,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.store.Save(link); err != nil {
		if errors.Is(err, storage.ErrCodeExists) {
			http.Error(w, "customAlias already in use", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to store link", http.StatusInternalServerError)
		return
	}

	shortURL := fmt.Sprintf("%s/%s", s.baseURL, code)
	writeJSON(w, http.StatusCreated, shortenResponse{
		Code:        code,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})
}

func (s *Server) handleListLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	links := s.store.List()
	items := make([]linkOverview, 0, len(links))
	for _, link := range links {
		items = append(items, linkOverview{
			Code:        link.Code,
			OriginalURL: link.OriginalURL,
			CreatedAt:   link.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleLinkDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasPrefix(r.URL.Path, "/api/links/") {
		http.NotFound(w, r)
		return
	}

	code := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/links/"), "/")
	if code == "" {
		http.NotFound(w, r)
		return
	}

	link, ok := s.store.Get(code)
	if !ok {
		http.NotFound(w, r)
		return
	}

	shortURL := fmt.Sprintf("%s/%s", s.baseURL, link.Code)
	resp := linkDetailsResponse{
		Code:        link.Code,
		ShortURL:    shortURL,
		OriginalURL: link.OriginalURL,
		CreatedAt:   link.CreatedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/" {
		http.NotFound(w, r)
		return
	}

	code := strings.Trim(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if code == "" {
		http.NotFound(w, r)
		return
	}

	link, ok := s.store.Get(code)
	if !ok {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, link.OriginalURL, http.StatusFound)
}
