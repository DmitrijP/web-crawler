package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
}

func normalizeURL(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("Empty string\n")
	}
	url = strings.Replace(url, "https://", "", -1)
	url = strings.Replace(url, "http://", "", -1)
	url = strings.TrimRight(url, "/")
	return url, nil
}

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	var links []string
	tokenizer := html.NewTokenizer(strings.NewReader(htmlBody))
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return nil, tokenizer.Err()
		}

		if tokenType == html.TokenType(html.StartTagToken) {
			token := tokenizer.Token()
			if "a" == token.Data {
				if token.Attr != nil && len(token.Attr) > 0 {
					for _, v := range token.Attr {
						if v.Key == "href" {
							if strings.HasPrefix(v.Val, "http") {
								links = append(links, v.Val)
								continue
							}
							if strings.HasPrefix(v.Val, "/") {
								links = append(links, rawBaseURL+v.Val)
							}
						}
					}
				}
			}
		}
	}
	return links, nil
}

func getHTML(rawURL string) (string, error) {
	res, err := http.Get(rawURL)
	if err != nil {
		fmt.Printf("Failed to load url: %v\n", err)
		return "", err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 400 {
		fmt.Printf("Response failed with status code: %d\n", res.StatusCode)
		return "", err
	}
	if err != nil {
		fmt.Printf("%v\n", err)
		return "", err
	}
	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		fmt.Printf("Wrong content-type %v.\n", res.Header)
		return "", err
	}
	return string(body), nil
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	defer cfg.wg.Done()
	defer func() { <-cfg.concurrencyControl }()
	cfg.concurrencyControl <- struct{}{}

	log.Printf("Crawling: %v\n", rawCurrentURL)
	rawCurrentURLNormalized, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error normalizing URL from %s, err: %v\n", rawCurrentURL, err)
		return
	}

	if !strings.Contains(rawCurrentURL, cfg.baseURL.String()) {
		fmt.Printf("Aborting, would leave %v to go to %v\n", cfg.baseURL.String(), rawCurrentURL)
		return
	}

	isFirst := cfg.addPageVisit(rawCurrentURLNormalized)
	if !isFirst {
		return
	}

	html, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error getting HTML from %s, err: %v\n", rawCurrentURL, err)
		return
	}

	urls, err := getURLsFromHTML(html, rawCurrentURL)
	if err != nil {
		fmt.Printf("Error getting URLs from %s, err: %v\n", rawCurrentURL, err)
		return
	}
	for _, v := range urls {
		cfg.wg.Add(1)
		go cfg.crawlPage(v)
	}
}

func crawlPage(rawBaseURL, rawCurrentURL string, pages map[string]int) {
	log.Printf("Crawling: %v\n", rawCurrentURL)
	rawCurrentURLNormalized, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error normalizing URL from %s, err: %v\n", rawCurrentURL, err)
		return
	}

	if !strings.Contains(rawCurrentURL, rawBaseURL) {
		fmt.Printf("Aborting, would leave %v to go to %v\n", rawBaseURL, rawCurrentURL)
		return
	}

	if _, ok := pages[rawCurrentURLNormalized]; ok {
		pages[rawCurrentURLNormalized]++
		return
	}
	pages[rawCurrentURLNormalized] = 1

	html, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("Error getting HTML from %s, err: %v\n", rawCurrentURL, err)
		return
	}

	urls, err := getURLsFromHTML(html, rawCurrentURL)
	if err != nil {
		fmt.Printf("Error getting URLs from %s, err: %v\n", rawCurrentURL, err)
		return
	}
	for _, v := range urls {
		crawlPage(rawCurrentURLNormalized, v, pages)
	}
}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if _, ok := cfg.pages[normalizedURL]; ok {
		cfg.pages[normalizedURL]++
		return false
	}
	cfg.pages[normalizedURL] = 1
	return true
}

func NewConfig(maxConcurrency int, baseURL *url.URL) *config {
	return &config{
		pages:              map[string]int{},
		baseURL:            baseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
	}
}
