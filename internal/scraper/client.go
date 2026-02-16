package scraper

import (
	"net/http"
	"time"
)

// Scraper — основная структура для работы с сайтом
type Scraper struct {
	client  *http.Client
	baseURL string
	lat     string
	lon     string
}

// NewScraper создает новый экземпляр
func NewScraper(baseURL, proxyAddr, lat, lon string) (*Scraper, error) {
	return &Scraper{
		baseURL: baseURL,
		lat:     lat,
		lon:     lon,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}, nil
}