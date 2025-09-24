package geocode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	// ErrAddressRequired is returned when no address is provided.
	ErrAddressRequired = errors.New("address is required")
	// ErrNoResults is returned when the Google API finds no results for the address.
	ErrNoResults = errors.New("no results found")
)

// Result represents a successful geocoding response.
type Result struct {
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Source    string  `json:"source"`
}

// Service performs geocoding requests against the Google Maps Geocoding API.
type Service struct {
	apiKey string
	client *http.Client
	cache  *cache
}

// NewService creates a configured Service instance. cacheTTL determines the lifetime of cache entries.
func NewService(apiKey string, cacheTTL time.Duration) *Service {
	return &Service{
		apiKey: apiKey,
		client: &http.Client{Timeout: 5 * time.Second},
		cache:  newCache(cacheTTL),
	}
}

// Geocode retrieves the coordinates for an address. It will use an in-memory cache before
// querying the Google Maps API to keep the service responsive under heavy load.
func (s *Service) Geocode(ctx context.Context, rawAddress string) (Result, error) {
	address := normalizeAddress(rawAddress)
	if address == "" {
		return Result{}, ErrAddressRequired
	}

	if result, ok := s.cache.Get(address); ok {
		result.Source = "cache"
		return result, nil
	}

	apiURL := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=%s",
		url.QueryEscape(address), s.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return Result{}, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("google maps api returned status %d", resp.StatusCode)
	}

	var payload geocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Result{}, err
	}

	if payload.Status != "OK" {
		if payload.ErrorMessage != "" {
			return Result{}, fmt.Errorf("google maps api error: %s", payload.ErrorMessage)
		}
		return Result{}, fmt.Errorf("google maps api status: %s", payload.Status)
	}

	if len(payload.Results) == 0 {
		return Result{}, ErrNoResults
	}

	top := payload.Results[0]
	result := Result{
		Address:   top.FormattedAddress,
		Latitude:  top.Geometry.Location.Lat,
		Longitude: top.Geometry.Location.Lng,
		Source:    "google",
	}

	s.cache.Set(address, result)

	return result, nil
}

func normalizeAddress(address string) string {
	return strings.TrimSpace(strings.ToLower(address))
}

// geocodeResponse models the subset of the Google Geocoding API response that we require.
type geocodeResponse struct {
	Results []struct {
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
}

// cache is a minimal in-memory cache with TTL support used to avoid expensive API calls for repeated requests.
type cache struct {
	ttl   time.Duration
	items map[string]cacheItem
	mu    sync.RWMutex
}

type cacheItem struct {
	value   Result
	expires time.Time
}

func newCache(ttl time.Duration) *cache {
	return &cache{
		ttl:   ttl,
		items: make(map[string]cacheItem),
	}
}

func (c *cache) Get(key string) (Result, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return Result{}, false
	}
	if time.Now().After(item.expires) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return Result{}, false
	}
	return item.value, true
}

func (c *cache) Set(key string, value Result) {
	c.mu.Lock()
	c.items[key] = cacheItem{
		value:   value,
		expires: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}
