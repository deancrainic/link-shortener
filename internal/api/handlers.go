package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
