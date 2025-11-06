package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
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

func fetch_page(Url string) (CrawlerMessage, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(Url)
	if err != nil {
		log.Printf("Error with code%d", err)
		return CrawlerMessage{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Page returned status code: %d", resp.StatusCode)
		return CrawlerMessage{}, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, Url)
	}

	message := CrawlerMessage{
		url: Url,
		meta: Metadata{
			status_code: resp.StatusCode,
		},
	}
	extractedData := extract_data(resp.Body)
	message.text = extractedData.text
	message.meta.title = extractedData.meta.title
	message.meta.description = extractedData.meta.description
	message.meta.timestamp = extractedData.meta.timestamp

	return message, nil
}

func extract_data(r io.Reader) CrawlerMessage {
	message := CrawlerMessage{}
	var res_text strings.Builder
	skip_tag := ""

	tokenizer := html.NewTokenizer(r)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
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
	message.meta.status_code = 200

	return message
}

func main() {
	log.Println("Crowler is running!")
	crawler_message, _ := fetch_page("https://example.com")
	fmt.Printf("🌐 URL: %s\n", crawler_message.url)
	fmt.Printf("📝 Title: %s\n", crawler_message.meta.title)
	fmt.Printf("📊 Status: %d\n", crawler_message.meta.status_code)
	fmt.Printf("⏰ Time: %s\n", crawler_message.meta.timestamp)
	fmt.Printf("📄 Text: %s\n", crawler_message.text[:100])
}
