package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type PageData struct {
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Links        []string  `json:"links"`
	Depth        int       `json:"depth"`
	CrawledAt    time.Time `json:"crawled_at"`
	ResponseTime int64     `json:"response_time_ms"`
	StatusCode   int       `json:"status_code"`
}

type CrawlResult struct {
	BaseURL     string     `json:"base_url"`
	MaxDepth    int        `json:"max_depth"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	TotalPages  int        `json:"total_pages"`
	Pages       []PageData `json:"pages"`
}

type Crawler struct {
	visited      map[string]bool
	visitedLock  sync.RWMutex
	baseURL      *url.URL
	maxDepth     int
	rateLimiter  <-chan time.Time
	result       CrawlResult
	resultLock   sync.Mutex
}

func NewCrawler(baseURL string, maxDepth int, requestsPerSecond float64) (*Crawler, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	return &Crawler{
		visited:     make(map[string]bool),
		baseURL:    parsedURL,
		maxDepth:   maxDepth,
		rateLimiter: time.Tick(time.Duration(1000/requestsPerSecond) * time.Millisecond),
		result: CrawlResult{
			BaseURL:   baseURL,
			MaxDepth:  maxDepth,
			StartTime: time.Now(),
			Pages:     make([]PageData, 0),
		},
	}, nil
}

func (c *Crawler) isVisited(url string) bool {
	c.visitedLock.RLock()
	defer c.visitedLock.RUnlock()
	return c.visited[url]
}

func (c *Crawler) markVisited(url string) {
	c.visitedLock.Lock()
	defer c.visitedLock.Unlock()
	c.visited[url] = true
}

func (c *Crawler) isSameDomain(pageURL *url.URL) bool {
	return pageURL.Host == c.baseURL.Host
}

func (c *Crawler) addPageData(data PageData) {
	c.resultLock.Lock()
	defer c.resultLock.Unlock()
	c.result.Pages = append(c.result.Pages, data)
}

func (c *Crawler) crawl(pageURL string, depth int, wg *sync.WaitGroup) {
	defer wg.Done()

	if depth > c.maxDepth {
		return
	}

	if c.isVisited(pageURL) {
		return
	}

	<-c.rateLimiter // Rate limiting

	c.markVisited(pageURL)
	fmt.Printf("Crawling: %s (depth: %d)\n", pageURL, depth)

	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		fmt.Printf("Error parsing URL %s: %v\n", pageURL, err)
		return
	}

	startTime := time.Now()
	resp, err := http.Get(pageURL)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", pageURL, err)
		return
	}
	defer resp.Body.Close()

	responseTime := time.Since(startTime).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: status code %d for %s\n", resp.StatusCode, pageURL)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing page %s: %v\n", pageURL, err)
		return
	}

	// Collect links
	links := make([]string, 0)
	doc.Find("a").Each(func(_ int, link *goquery.Selection) {
		href, exists := link.Attr("href")
		if !exists {
			return
		}

		href = strings.TrimSpace(href)
		if href == "" || strings.HasPrefix(href, "#") {
			return
		}

		absoluteURL, err := parsedURL.Parse(href)
		if err != nil {
			return
		}

		if !c.isSameDomain(absoluteURL) {
			return
		}

		nextURL := absoluteURL.String()
		links = append(links, nextURL)

		if !c.isVisited(nextURL) {
			wg.Add(1)
			go c.crawl(nextURL, depth+1, wg)
		}
	})

	// Create and store page data
	pageData := PageData{
		URL:          pageURL,
		Title:        doc.Find("title").Text(),
		Links:        links,
		Depth:        depth,
		CrawledAt:    time.Now(),
		ResponseTime: responseTime,
		StatusCode:   resp.StatusCode,
	}

	c.addPageData(pageData)
}

func (c *Crawler) saveResults(filename string) error {
	c.result.EndTime = time.Now()
	c.result.TotalPages = len(c.result.Pages)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c.result); err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}

	return nil
}

func (c *Crawler) Start() error {
	var wg sync.WaitGroup
	wg.Add(1)
	go c.crawl(c.baseURL.String(), 0, &wg)
	wg.Wait()

	fmt.Printf("\nCrawling completed. Total pages visited: %d\n", len(c.visited))
	
	// Save results to JSON file
	err := c.saveResults("crawl_results.json")
	if err != nil {
		return fmt.Errorf("error saving results: %v", err)
	}

	fmt.Println("Results saved to crawl_results.json")
	return nil
}

func main() {

	fmt.Println("Starting crawler... \n Enter the base URL: ")
	
	var baseURL string
	fmt.Scanln(&baseURL)
	
	maxDepth := 3
	requestsPerSecond := 2.0

	crawler, err := NewCrawler(baseURL, maxDepth, requestsPerSecond)
	if err != nil {
		fmt.Printf("Error creating crawler: %v\n", err)
		return
	}

	if err := crawler.Start(); err != nil {
		fmt.Printf("Error during crawling: %v\n", err)
		return
	}
}