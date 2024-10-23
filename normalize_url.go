package main

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

func normalizeURL(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("Empty string")
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
				fmt.Printf("Attr: %v", token.Attr)
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
