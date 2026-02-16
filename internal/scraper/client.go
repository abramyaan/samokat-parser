package scraper

import (
	"net/http"
	"net/url"
	"time"
)

type Scraper struct {
	client *http.Client
	baseURL string
}

func NewScraper(baseURL string, proxyAddr string)(*Scraper,error){
	transport := &http.Transport{}
	if proxyAddr != "" {
		proxyURL, err := url.Parse(proxyAddr)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &Scraper{
		baseURL: baseURL,
		client: &http.Client{
			Transport: transport,
			Timeout: 30 * time.Second,
		},
	}, nil
}