package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Metadata struct {
	title       string
	description string
	timestamp   string
	status_code int
}

type CrawlerMessage struct {
	url  string
	text string
	meta Metadata
}

func attrToMap(arr_attr []html.Attribute) map[string]string {
	map_attr := make(map[string]string, len(arr_attr))
	for _, elem := range arr_attr {
		map_attr[elem.Key] = elem.Val
	}
	return map_attr
}

func valueExistAndEqualKey(attr_map map[string]string, key, value string) bool {
	if val, ok := attr_map[key]; ok {
		if val == value {
			return true
		}
	}
	return false
}

func normalizeUrl(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err.Error()
	}
	parsedURL.Fragment = ""
	parsedURL.Host = strings.ToLower(parsedURL.Host)
	parsedURL.Host = strings.TrimPrefix(parsedURL.Host, "www.")
	return parsedURL.String()
}

func normalizeText(text string) string {
	if text == "" {
		return ""
	}

	text = strings.ToLower(text)
	text = regexp.MustCompile(`\n+`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}

func toAbsolute(base string, rawlink string) string {

	baseParsed, err := url.Parse(base)
	if err != nil {
		return ""
	}

	rawLinkParsed, err := url.Parse(rawlink)
	if err != nil {
		return ""
	}

	return baseParsed.ResolveReference(rawLinkParsed).String()
}

func isValidLink(Url string, link string) (string, bool) {

	if link == "" || link == "#" {
		return "", false
	}

	abs := toAbsolute(Url, link)

	abs = normalizeUrl(abs)

	if !strings.HasPrefix(abs, "http://") && !strings.HasPrefix(abs, "https://") {
		return "", false
	}

	return abs, true
}

func extractData(r io.Reader) (CrawlerMessage, []string) {
	message := CrawlerMessage{}
	var next_links []string
	var res_text strings.Builder
	skip_tag := ""

	tokenizer := html.NewTokenizer(r)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			if err := tokenizer.Err(); err != io.EOF {
				log.Printf("HTML parse error: %v", err)
			}
			break
		}

		if skip_tag != "" {
			if tt == html.EndTagToken && tokenizer.Token().Data == skip_tag {
				skip_tag = ""
			}
			continue
		}

		switch tt {

		case html.StartTagToken, html.SelfClosingTagToken:
			tok := tokenizer.Token()

			switch tok.Data {
			case "title":
				if tokenizer.Next() == html.TextToken {
					message.meta.title = tokenizer.Token().Data
				}

			case "meta":
				attr_map := attrToMap(tok.Attr)
				if valueExistAndEqualKey(attr_map, "property", "og:description") {
					message.meta.description = attr_map["content"]
				} else if valueExistAndEqualKey(attr_map, "name", "description") && !valueExistAndEqualKey(attr_map, "property", "og:description") {
					message.meta.description = attr_map["content"]
				}

			case "script", "noscript", "style":
				skip_tag = tok.Data

			case "h1", "h2", "h3", "h4", "h5", "h6":
				res_text.WriteString("\n")

			case "p", "div", "br":
				res_text.WriteString("\n")

			case "li":
				res_text.WriteString("\n• ")

			case "a":
				attr_map := attrToMap(tok.Attr)
				if link, ok := attr_map["href"]; ok {
					next_links = append(next_links, link)
				}
			}

		case html.TextToken:

			text := strings.TrimSpace(tokenizer.Token().Data)
			if text != "" {
				res_text.WriteString(text)
				res_text.WriteString(" ")
			}
		}

	}

	message.text = normalizeText(res_text.String())
	message.meta.timestamp = time.Now().UTC().Format(time.RFC3339)
	return message, next_links
}

func fetchPage(pageUrl string) (CrawlerMessage, []string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(pageUrl)
	if err != nil {
		log.Printf("Error with code%v", err)
		return CrawlerMessage{}, nil, err
	}
	defer resp.Body.Close()

	log.Printf("Page returned status code: %d", resp.StatusCode)
	message := CrawlerMessage{
		url: pageUrl,
		meta: Metadata{
			status_code: resp.StatusCode,
		},
	}
	extractedData, next_links := extractData(resp.Body)
	message.text = extractedData.text
	message.meta.title = extractedData.meta.title
	message.meta.description = extractedData.meta.description
	message.meta.timestamp = extractedData.meta.timestamp

	var validLinks []string
	isAlreadyAdded := make(map[string]bool)
	isAlreadyAdded[pageUrl] = true
	for _, link := range next_links {
		if normalizedLink, ok := isValidLink(pageUrl, link); ok {
			if !isAlreadyAdded[normalizedLink] {
				validLinks = append(validLinks, normalizedLink)
				isAlreadyAdded[normalizedLink] = true
			}
		}
	}
	time.Sleep(200 * time.Millisecond)

	return message, validLinks, nil
}

func startCrawler(Url string) error {
	linkQueue := []string{normalizeUrl(Url)}
	visited := make(map[string]bool)

	for i := 0; i < len(linkQueue); i++ {
		url := linkQueue[i]

		if !visited[url] {
			message, nextLinks, err := fetchPage(url)
			if err != nil {
				log.Printf("fetch page error: url = %s, err = %v", url, err)
				continue
			}
			fmt.Printf("🌐 URL: %s\n", message.url)
			fmt.Printf("📝 Title: %s\n", message.meta.title)
			fmt.Printf("📊 Status: %d\n", message.meta.status_code)
			fmt.Printf("⏰ Time: %s\n", message.meta.timestamp)
			fmt.Printf("📄 Text: %s\n", message.text)

			visited[url] = true
			for l := 0; l < len(nextLinks); l++ {
				if !visited[nextLinks[l]] {
					linkQueue = append(linkQueue, nextLinks[l])
				}
			}
		}
	}

	log.Printf("The page with url = %s is fully crawled", Url)
	return nil
}

func main() {
	log.Println("Crawler is running!")
	err := startCrawler("https://example.com")
	if err != nil {
		log.Printf(" %d", err)
	}
}
