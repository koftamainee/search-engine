package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meilisearch/meilisearch-go"

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
		"Портал:",
		"Portal:",
		"Шаблон:",
		"Template:",
		"Категория:",
		"Category:",
		"Файл:",
		"File:",
	}

	lowerURL := strings.ToLower(rawURL)
	for _, pattern := range unwantedPatterns {
		if strings.Contains(lowerURL, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
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
		return ""
	}
	parsedURL.Fragment = ""
	parsedURL.Host = strings.ToLower(parsedURL.Host)
	parsedURL.Host = strings.TrimPrefix(parsedURL.Host, "www.")

	if parsedURL.Scheme == "http" {
		parsedURL.Scheme = "https"
	}

	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/")
	if parsedURL.Path == "" {
		parsedURL.Path = "/"
	}
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

	if len(text) > 5000 {
		text = text[:5000]
	}
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

func isValidLink(ctx context.Context, rdb *redis.Client, baseURL string, link string) (string, bool) {

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

	allowed, _ := isAllowByRobots(ctx, rdb, abs, getDomain(abs))
	if !allowed {
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

const robotsCachePrefix = "robots:"
const robotsCacheTTL = 96 * time.Hour

// returns [ nil, nil ] if not in cache
func getRobotsFromCache(ctx context.Context, rdb *redis.Client, domain string) (*RobotsTxt, error) {
	val, err := rdb.Get(ctx, robotsCachePrefix+domain).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var robots RobotsTxt
	if err := json.Unmarshal([]byte(val), &robots); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &robots, nil
}

func setRobotsToCache(ctx context.Context, rdb *redis.Client, robots *RobotsTxt) error {
	data, err := json.Marshal(robots)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	return rdb.Set(ctx, robotsCachePrefix+robots.PageDomain, data, robotsCacheTTL).Err()
}

func fetchRobotsTxt(ctx context.Context, rdb *redis.Client, pageUrl string) (*RobotsTxt, error) {
	parsedURL, err := url.Parse(pageUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	robotTxtRef := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(robotTxtRef)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch robots.txt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		robots := &RobotsTxt{
			PageDomain:    parsedURL.Host,
			DisallowPaths: make(map[string]bool),
		}
		if err := setRobotsToCache(ctx, rdb, robots); err != nil {
			log.Printf("failed to cache empty robots.txt for %s: %v", parsedURL.Host, err)
		}
		return robots, nil
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		log.Printf("robots.txt for %s returned %d, treating as full disallow", parsedURL.Host, resp.StatusCode)
		robots := &RobotsTxt{
			PageDomain:    parsedURL.Host,
			DisallowPaths: map[string]bool{"/": true},
		}
		if err := setRobotsToCache(ctx, rdb, robots); err != nil {
			log.Printf("failed to cache full-disallow robots.txt for %s: %v", parsedURL.Host, err)
		}
		return robots, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for robots.txt", resp.StatusCode)
	}

	robots := &RobotsTxt{
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
			agentLower := strings.ToLower(agent)
			isRelevantUserAgent = (agentLower == "*" || agentLower == "skwajer's_searchenginecrawler/1.0")
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
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading robots.txt: %w", err)
	}

	if err := setRobotsToCache(ctx, rdb, robots); err != nil {
		log.Printf("failed to cache robots.txt for %s: %v", parsedURL.Host, err)
	}

	return robots, nil
}

func CacheByRobots(ctx context.Context, rdb *redis.Client, pageUrl string) {
	domain := getDomain(pageUrl)
	robots, err := getRobotsFromCache(ctx, rdb, domain)
	if err != nil {
		log.Printf("error checking robots cache for %s: %v", domain, err)
		return
	}
	if robots == nil {
		_, err := fetchRobotsTxt(ctx, rdb, pageUrl)
		if err != nil {
			log.Printf("failed to fetch robots.txt for %s: %v", domain, err)
		}
	}
}

func isAllowByRobots(ctx context.Context, rdb *redis.Client, rawURL string, domain string) (bool, error) {
	robots, err := getRobotsFromCache(ctx, rdb, domain)
	if err != nil {
		return true, err
	}

	if robots == nil {

		return true, nil
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return true, nil
	}

	path := parsed.Path

	for disallowedPath := range robots.DisallowPaths {
		if strings.HasPrefix(path, disallowedPath) {
			return false, nil
		}
	}

	return true, nil
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

func fetchPage(ctx context.Context, rdb *redis.Client, pageUrl string) (CrawlerMessage, []string, error) {
	CacheByRobots(ctx, rdb, pageUrl)

	allowed, _ := isAllowByRobots(ctx, rdb, pageUrl, getDomain(pageUrl))
	if !allowed {
		log.Printf("%s is disallowed by robots.txt", pageUrl)
		return CrawlerMessage{}, nil, errors.New("page is disallowed by robots.txt")
	}

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
	isAlreadyAdded := make(map[string]bool)
	isAlreadyAdded[pageUrl] = true

	for _, link := range next_links {
		select {
		case <-ctx.Done():
			return message, validLinks, ctx.Err()
		default:
		}

		if normalizedLink, ok := isValidLink(ctx, rdb, pageUrl, link); ok {
			if isUnwantedURL(normalizedLink) {
				continue
			}
			if !isAlreadyAdded[normalizedLink] {
				validLinks = append(validLinks, normalizedLink)
				isAlreadyAdded[normalizedLink] = true
			}
		}
	}

	return message, validLinks, nil
}

func startCrawler(ctx context.Context, rdb *redis.Client, meiliIndex meilisearch.IndexManager, startUrl string, numWorkers int) error {
	normalized := normalizeUrl(startUrl)
	added, err := rdb.SAdd(ctx, "visited", normalized).Result()
	if err != nil {
		return errors.New("Page already crawled")
	}
	if added != 0 {
		rdb.LPush(ctx, "queue", normalized)
	}

	urlChan := make(chan string, 100)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					log.Printf("[worker %d] shutting down...", workerID)
					return
				case url, ok := <-urlChan:
					if !ok {
						return
					}

					reqCtx, cancelReq := context.WithTimeout(ctx, 10*time.Second)
					message, nextLinks, err := fetchPage(reqCtx, rdb, url)
					cancelReq()

					if err != nil {
						if ctx.Err() != nil {
							return
						}
						log.Printf("[worker %d] fetch error: url=%s err=%v", workerID, url, err)
						continue
					}

					fmt.Printf("[worker %d] 🌐 URL: %s\n", workerID, message.Url)
					fmt.Printf("[worker %d] 📝 Title: %s\n", workerID, message.Meta.Title)
					fmt.Printf("[worker %d] 📊 Status: %d\n", workerID, message.Meta.Status_code)

					id := fmt.Sprintf("%x", md5.Sum([]byte(message.Url)))

					if message.Meta.Status_code == 200 {
						docs := []map[string]interface{}{
							{
								"id":          id,
								"url":         message.Url,
								"title":       message.Meta.Title,
								"description": message.Meta.Description,
								"text":        message.Text,
								"timestamp":   message.Meta.Timestamp,
							},
						}
						task, err := meiliIndex.AddDocuments(docs, nil)
						if err != nil {
							log.Printf("[worker %d] Meilisearch index error: %v", workerID, err)
							continue
						}
						fmt.Printf("[worker %d] 📑 Indexed to Meilisearch (task: %d)\n", workerID, task.TaskUID)
					}

					for _, nextLink := range nextLinks {
						select {
						case <-ctx.Done():
							return
						default:
						}

						norm := normalizeUrl(nextLink)
						added, err := rdb.SAdd(ctx, "visited", norm).Result()
						if err != nil {
							continue
						}
						if added != 0 {
							rdb.LPush(ctx, "queue", norm)
						}
					}
				}
			}
		}(i)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			url, err := rdb.RPop(ctx, "queue").Result()
			if err == redis.Nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(300 * time.Millisecond):
					continue
				}
			}
			if err != nil {
				log.Printf("Error reading from queue: %v", err)
				return
			}

			select {
			case <-ctx.Done():
				rdb.LPush(ctx, "queue", url)
				return
			case urlChan <- url:
			}
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down crawler...")

	close(urlChan)

	wg.Wait()

	log.Println("All workers stopped")
	return ctx.Err()

}

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

	//health check endpoint
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		server := &http.Server{
			Addr:    ":8081",
			Handler: mux,
		}
		log.Println("Health check server listening on :8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health server error: %v", err)
		}

	}()

	//redis initialization
	host := os.Getenv("REDIS_CRAWLER_HOST")
	port := os.Getenv("REDIS_CRAWLER_PORT")
	password := os.Getenv("REDIS_CRAWLER_PASSWORD")

	addr := fmt.Sprintf("%s:%s", host, port)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		fmt.Println("Redis is not working:", err)
		return
	}
	fmt.Println("Redis is working.")

	// Meilisearch initialization
	meiliHost := os.Getenv("MEILI_HOST")
	meiliPort := os.Getenv("MEILI_PORT")
	meiliMasterKey := os.Getenv("MEILI_MASTER_KEY")

	meiliURL := fmt.Sprintf("http://%s:%s", meiliHost, meiliPort)
	meiliClient := meilisearch.New(meiliURL, meilisearch.WithAPIKey(meiliMasterKey))

	_, err = meiliClient.Health()
	if err != nil {
		fmt.Printf("Meilisearch connection error: %v\n", err)
		return
	}
	fmt.Println("Meilisearch is working.")
	fmt.Printf("Meilisearch URL: %s\n", meiliURL)

	meiliIndex := meiliClient.Index("web_pages")

	numWorkers := 15

	if err := startCrawler(ctx, rdb, meiliIndex, "https://example.com", numWorkers); err != nil {
		if err == context.Canceled {
			log.Println("Crawler stopped by user")
		} else {
			log.Printf("Crawler error: %v", err)
		}
	}
}
