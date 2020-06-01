package stockings

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

const urlFmt = "https://robinhood.com/stocks/"

// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (ok bool, href string) {
	// Iterate over token attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	// "bare" return will return the variables (ok, href) as
	// defined in the function definition
	return
}

func ScrapeCollections(ticker string) ([]string, error) {
	resp, err := http.Get(fmt.Sprint(urlFmt, ticker))
	if err != nil {
		return nil, err
	}
	z := html.NewTokenizer(resp.Body)
	defer resp.Body.Close()

	collectionMap := make(map[string]bool)
L:
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			break L
		case tt == html.StartTagToken:
			t := z.Token()

			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract the href value, if there is one
			ok, url := getHref(t)
			if !ok {
				continue
			}
			if strings.Contains(url, "/collections/") {
				collectionMap[url] = true
			}
		}
	}
	collections := make([]string, 0)
	for url := range collectionMap {
		urlName := strings.ReplaceAll(url, "/collections/", "")
		collections = append(collections, urlName)
	}
	return collections, nil
}
