package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"link-shortener/internal/shortcode"
	"link-shortener/internal/storage"
)

var errInvalidCustomCode = errors.New("customAlias must be 3-30 characters (letters, numbers, underscores, hyphens)")

func (s *Server) resolveCode(custom string) (string, error) {
	code := strings.TrimSpace(custom)
	if code != "" {
		if !codePattern.MatchString(code) {
			return "", errInvalidCustomCode
		}
		if _, exists := s.store.Get(code); exists {
			return "", storage.ErrCodeExists
		}
		return code, nil
	}
	return s.generateUniqueCode()
}

func (s *Server) generateUniqueCode() (string, error) {
	for attempts := 0; attempts < 5; attempts++ {
		code, err := shortcode.Generate(minCodeLength, maxCodeLength)
		if err != nil {
			return "", err
		}
		if _, exists := s.store.Get(code); !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("unable to find unique code after several attempts")
}

func validateURL(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("url is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("url must start with http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("url must include host")
	}
	return parsed.String(), nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

var codePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
