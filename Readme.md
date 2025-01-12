# Concurrent Web Crawler in Go

A concurrent web crawler written in Go that systematically crawls websites while respecting rate limits and domain boundaries. The crawler collects page data and saves the results in a structured JSON format.

## Features

1. **Concurrent Crawling**: 
   - Utilizes Go's goroutines for efficient parallel crawling.
   
2. **Rate Limiting**: 
   - Configurable requests per second to avoid overwhelming target servers.
   
3. **Domain Boundary Respect**: 
   - Only crawls pages within the same domain.
   
4. **Depth Control**: 
   - Configurable maximum crawl depth.
   
5. **Data Collection**: 
   - Captures detailed information about each crawled page, including title, links, response time, and HTTP status.
   
6. **JSON Output**: 
   - Saves crawl results in a structured JSON format.

## Installation

1. Ensure you have [Go](https://golang.org/) installed on your system.
2. Clone this repository.
3. Install dependencies:

```bash
go get github.com/PuerkitoBio/goquery
```

## Usage

1. Run the crawler:

```bash
go run main.go
```

2. Follow the prompts to enter the base URL and configure other parameters.

The crawler will start with the following default settings:

Maximum Depth: 3 levels
Rate Limit: 2 requests per second
Sample Usage
```bash
$ go run main.go

Starting crawler... 
Enter the base URL:
https://example.com

Crawling: https://example.com (depth: 0)
Crawling: https://example.com/about (depth: 1)
Crawling: https://example.com/products (depth: 1)
Crawling: https://example.com/contact (depth: 1)
Crawling: https://example.com/products/item1 (depth: 2)
Crawling: https://example.com/products/item2 (depth: 2)

Crawling completed. Total pages visited: 6
Results saved to crawl_results.json
```
### Sample Output

Here's an example of the generated crawl_results.json:

```bash
{
  "base_url": "https://example.com",
  "max_depth": 3,
  "start_time": "2025-01-12T10:30:00Z",
  "end_time": "2025-01-12T10:30:05Z",
  "total_pages": 6,
  "pages": [
    {
      "url": "https://example.com",
      "title": "Example Website",
      "links": [
        "https://example.com/about",
        "https://example.com/products",
        "https://example.com/contact"
      ],
      "depth": 0,
      "crawled_at": "2025-01-12T10:30:00Z",
      "response_time_ms": 123,
      "status_code": 200
    },
    {
      "url": "https://example.com/products",
      "title": "Products - Example Website",
      "links": [
        "https://example.com/products/item1",
        "https://example.com/products/item2"
      ],
      "depth": 1,
      "crawled_at": "2025-01-12T10:30:02Z",
      "response_time_ms": 95,
      "status_code": 200
    }
  ]
}
```