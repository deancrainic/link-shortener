package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	qrcode "github.com/skip2/go-qrcode"

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

func parseExpiresAt(raw *string) (time.Time, error) {
	now := time.Now().UTC()
	defaultExpiry := now.Add(30 * 24 * time.Hour)
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return defaultExpiry, nil
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw))
	if err != nil {
		return time.Time{}, errors.New("expiresAt must be RFC3339 timestamp")
	}
	if t.Before(now) {
		return time.Time{}, errors.New("expiresAt must be in the future")
	}
	return t.UTC(), nil
}

var (
	codePattern        = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
	defaultGeoEndpoint = "https://ipapi.co/%s/country/"
	geoCacheTTL        = time.Hour
	geoHTTPClient      = &http.Client{Timeout: 3 * time.Second}
	geoCache           = struct {
		sync.Mutex
		data map[string]geoCacheEntry
	}{data: make(map[string]geoCacheEntry)}
)

type geoCacheEntry struct {
	country string
	expires time.Time
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func generateQRCodeDataURL(content string) (string, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(png)), nil
}

func detectCountry(ip string) string {
	if ip == "" {
		return "Unknown"
	}
	if country, ok := cachedCountry(ip); ok {
		return country
	}
	if country, err := fetchCountryByIP(ip); err == nil && country != "" {
		storeCountry(ip, country)
		return country
	}
	return "Unknown"
}

func cachedCountry(ip string) (string, bool) {
	geoCache.Lock()
	defer geoCache.Unlock()

	entry, ok := geoCache.data[ip]
	if !ok {
		return "", false
	}
	if time.Now().After(entry.expires) {
		delete(geoCache.data, ip)
		return "", false
	}
	return entry.country, true
}

func storeCountry(ip, country string) {
	if country == "" {
		return
	}
	geoCache.Lock()
	defer geoCache.Unlock()
	geoCache.data[ip] = geoCacheEntry{
		country: country,
		expires: time.Now().Add(geoCacheTTL),
	}
}

func fetchCountryByIP(ip string) (string, error) {
	endpoint := geoEndpoint()
	var lookupURL string
	if strings.Contains(endpoint, "%s") {
		lookupURL = fmt.Sprintf(endpoint, url.PathEscape(ip))
	} else {
		lookupURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(endpoint, "/"), url.PathEscape(ip))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lookupURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := geoHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("geo lookup failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	country := strings.TrimSpace(string(body))
	if country == "" {
		return "", errors.New("empty geo response")
	}
	return country, nil
}

func geoEndpoint() string {
	if val := strings.TrimSpace(os.Getenv("GEOIP_ENDPOINT")); val != "" {
		return val
	}
	return defaultGeoEndpoint
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
