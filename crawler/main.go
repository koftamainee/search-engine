package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html"
)

type Metadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Timestamp   string `json:"timestamp,omitempty"`
	Status_code int    `json:"statusCode"`
}

type CrawlerMessage struct {
	Url  string   `json:"url"`
	Text string   `json:"text"`
	Meta Metadata `json:"meta"`
}

type RobotsTxt struct {
	PageDomain    string
	DisallowPaths map[string]bool
}

func isUnwantedURL(rawURL string) bool {
	unwantedPatterns := []string{
		"action=edit",
		"action=history",
		"action=info",
		"veaction=edit",
		"oldid=",
		"diff=",
		"printable=yes",
		"redlink=1",
		"Служебная:",
		"Special:",
		"Справка:",
		"Help:",
		"Википедия:",
		"Wikipedia:",
		"Портал:",
		"Portal:",
		"Шаблон:",
		"Template:",
		"Категория:",
		"Category:",
		"Файл:",
		"File:",
		"Обсуждение:",
		"Talk:",
	}

	lowerURL := strings.ToLower(rawURL)
	for _, pattern := range unwantedPatterns {
		if strings.Contains(lowerURL, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// TODO: make redis cache here
var RobotsCache = make(map[string]*RobotsTxt)

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
		return ""
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
	if strings.HasPrefix(rawlink, "http://") || strings.HasPrefix(rawlink, "https://") {
		return rawlink
	}

	baseParsed, err := url.Parse(base)
	if err != nil {
		return ""
	}

	if strings.HasPrefix(rawlink, "/") {
		return baseParsed.Scheme + "://" + baseParsed.Host + rawlink
	}

	rawLinkParsed, err := url.Parse(rawlink)
	if err != nil {
		return ""
	}

	resolved := baseParsed.ResolveReference(rawLinkParsed)
	return resolved.String()
}

func isValidLink(baseURL string, link string) (string, bool) {

	if link == "" || link == "#" {
		return "", false
	}

	if strings.HasPrefix(link, "javascript:") ||
		strings.HasPrefix(link, "mailto:") ||
		strings.HasPrefix(link, "tel:") {
		return "", false
	}

	abs := toAbsolute(baseURL, link)
	if abs == "" {
		return "", false
	}

	abs = normalizeUrl(abs)
	if abs == "" {
		return "", false
	}

	if !strings.HasPrefix(abs, "http://") && !strings.HasPrefix(abs, "https://") {
		return "", false
	}

	if !isAllowByRobots(abs, getDomain(abs)) {
		return "", false
	}

	return abs, true
}

func getDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := parsed.Hostname()
	host = strings.TrimPrefix(host, "www.")
	return host
}

func fetchRobotsTxt(pageUrl string) (RobotsTxt, error) {
	parsedURL, err := url.Parse(pageUrl)
	if err != nil {
		return RobotsTxt{}, fmt.Errorf("failed to parse URL: %w", err)
	}

	robotTxtRef := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(robotTxtRef)
	if err != nil {
		return RobotsTxt{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return RobotsTxt{}, nil
	}

	robots := RobotsTxt{
		PageDomain:    parsedURL.Host,
		DisallowPaths: make(map[string]bool),
	}

	scanner := bufio.NewScanner(resp.Body)
	isRelevantUserAgent := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lowerLine := strings.ToLower(line)

		if strings.HasPrefix(lowerLine, "user-agent:") {
			agent := strings.TrimSpace(line[len("user-agent:"):])
			isRelevantUserAgent = (agent == "*" || agent == "search-engine")
			continue
		}

		if !isRelevantUserAgent {
			continue
		}

		if strings.HasPrefix(lowerLine, "disallow:") {
			path := strings.TrimSpace(line[len("disallow:"):])
			if path != "" {
				robots.DisallowPaths[path] = true
			}
		} else if strings.HasPrefix(lowerLine, "allow:") {
			path := strings.TrimSpace(line[len("allow:"):])
			if path != "" {
				delete(robots.DisallowPaths, path)
			}
		} else if strings.HasPrefix(lowerLine, "crawl-delay:") {
			continue
		} else if strings.HasPrefix(lowerLine, "sitemap:") {
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return RobotsTxt{}, fmt.Errorf("error reading robots.txt: %w", err)
	}

	RobotsCache[parsedURL.Host] = &robots
	return robots, nil
}

func CacheByRobots(PageUrl string) {
	domain := getDomain(PageUrl)
	_, ok := RobotsCache[domain]
	if !ok {
		fetchRobotsTxt(PageUrl)
	}
}

func isAllowByRobots(rawURL string, domain string) bool {
	robots, ok := RobotsCache[domain]
	if !ok {
		return true
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return true
	}

	path := parsed.Path

	for disallowedPath := range robots.DisallowPaths {
		if strings.HasPrefix(path, disallowedPath) {
			return false
		}
	}

	return true
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
					message.Meta.Title = tokenizer.Token().Data
				}

			case "meta":
				attr_map := attrToMap(tok.Attr)

				if valueExistAndEqualKey(attr_map, "property", "og:description") {
					message.Meta.Description = attr_map["content"]
				} else if valueExistAndEqualKey(attr_map, "name", "description") {
					if message.Meta.Description == "" {
						message.Meta.Description = attr_map["content"]
					}
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

	message.Text = normalizeText(res_text.String())
	message.Meta.Timestamp = time.Now().UTC().Format(time.RFC3339)
	return message, next_links
}

func fetchPage(ctx context.Context, pageUrl string) (CrawlerMessage, []string, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", pageUrl, nil)
	if err != nil {
		return CrawlerMessage{}, nil, err
	}

	req.Header.Set("User-Agent", "Skwajer's_SearchEngineCrawler/1.0")
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			log.Printf("context canceled request")
		}
		return CrawlerMessage{}, nil, err
	}
	defer resp.Body.Close()

	CacheByRobots(pageUrl)

	log.Printf("Page returned status code: %d", resp.StatusCode)
	message := CrawlerMessage{
		Url: pageUrl,
		Meta: Metadata{
			Status_code: resp.StatusCode,
		},
	}
	extractedData, next_links := extractData(resp.Body)
	message.Text = extractedData.Text
	message.Meta.Title = extractedData.Meta.Title
	message.Meta.Description = extractedData.Meta.Description
	message.Meta.Timestamp = extractedData.Meta.Timestamp

	var validLinks []string
	var externalLinks []string
	var internalLinks []string

	isAlreadyAdded := make(map[string]bool)
	isAlreadyAdded[pageUrl] = true
	for _, link := range next_links {

		select {
		case <-ctx.Done():
			return message, validLinks, ctx.Err()
		default:
		}

		if normalizedLink, ok := isValidLink(pageUrl, link); ok {
			if isUnwantedURL(normalizedLink) {
				continue
			}
			if !isAlreadyAdded[normalizedLink] {
				if getDomain(normalizedLink) != getDomain(pageUrl) {
					externalLinks = append(externalLinks, normalizedLink)
				} else {
					internalLinks = append(internalLinks, normalizedLink)
				}
				isAlreadyAdded[normalizedLink] = true
			}
		}
	}
	validLinks = append(validLinks, externalLinks...)
	validLinks = append(validLinks, internalLinks...)

	select {
	case <-time.After(200 * time.Millisecond):
	case <-ctx.Done():
		return message, validLinks, ctx.Err()
	}

	return message, validLinks, nil
}

func startCrawler(ctx context.Context, Url string) error {
	linkQueue := []string{normalizeUrl(Url)}
	visited := make(map[string]bool)

	for i := 0; i < len(linkQueue); i++ {

		if ctx.Err() != nil {
			log.Printf("Crawler stopped: %v", ctx.Err())
			return ctx.Err()
		}

		url := linkQueue[i]

		if !visited[url] {
			message, nextLinks, err := fetchPage(ctx, url)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				log.Printf("fetch page error: url = %s, err = %v", url, err)
				continue
			}
			fmt.Printf("🌐 URL: %s\n", message.Url)
			fmt.Printf("📝 Title: %s\n", message.Meta.Title)
			fmt.Printf("📊 Status: %d\n", message.Meta.Status_code)
			fmt.Printf("⏰ Time: %s\n", message.Meta.Timestamp)
			//fmt.Printf("📄 Text: %s\n", message.Text)

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

// TODO: i don't now were yet, but handle multi-thread processing
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received interrupt signal. Shutting down...")
		cancel()
	}()

	//redis initialization
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "!$bibleTumbSha256$!",
		DB:       0,
	})

	err := client.Ping().Err()
	if err != nil {
		fmt.Println("Redis is not working:", err)
		return
	}
	fmt.Println("Redis is working.")

	if err := startCrawler(ctx, "https://en.wikipedia.org/wiki/Main_Page"); err != nil {
		if err == context.Canceled {
			log.Println("Crawler stopped by user")
		} else {
			log.Printf("Crawler error: %v", err)
		}
	}
}
