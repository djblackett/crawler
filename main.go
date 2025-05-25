package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	}

	var maxConcurrency int
	var maxPages int
	var err error
	if len(args) > 1 {

		maxConcurrency, err = strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("invalid value for maxConcurrency", err)
			os.Exit(1)
		}
	} else {
		maxConcurrency = 5
	}

	if len(args) > 2 {
		maxPages, err = strconv.Atoi(args[2])
		if err != nil {
			fmt.Println("invalid value for maxPages", err)
			os.Exit(1)
		}
	} else {
		maxPages = 10
	}

	baseURL := args[0]
	parsedUrl, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg := config{
		pages:              make(map[string]int),
		baseURL:            parsedUrl,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPages,
	}
	fmt.Println("starting crawl of:", baseURL)

	cfg.wg.Add(1)
	go cfg.crawlPage(baseURL)

	cfg.wg.Wait()

	// fmt.Println(cfg.pages)
	printReport(cfg.pages, baseURL)

}

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
}

func getHTML(rawURL string) (string, error) {
	res, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode > 399 {
		return "", errors.New("received error response")
	}

	if !strings.HasPrefix(res.Header.Get("Content-Type"), "text/html") {
		fmt.Println("content-type:", res.Header.Get("Content-Type"))
		return "", nil
	}

	text, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(text), err
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	defer cfg.wg.Done()

	cfg.concurrencyControl <- struct{}{}
	defer func() { <-cfg.concurrencyControl }()

	normalizedURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Println("error normalizing currentUrl:", err)
		return
	}

	cfg.mu.Lock()
	if len(cfg.pages) >= cfg.maxPages {
		cfg.mu.Unlock()
		return
	}
	if _, seen := cfg.pages[normalizedURL]; seen {
		cfg.mu.Unlock()
		return
	}
	cfg.pages[normalizedURL] = 1
	cfg.mu.Unlock()

	current, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Println("error parsing currentBaseUrl")
		return
	}

	if current.Hostname() != cfg.baseURL.Hostname() {
		return
	}

	// isFirst := cfg.addPageVisit(normalizedURL)
	// if !isFirst {
	// 	return
	// }

	html, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Println("error getting html:", err)
		return
	}
	fmt.Println(html)

	urls, err := getURLsFromHTML(html, cfg.baseURL)
	if err != nil {
		fmt.Println("error getting urls from html", err)
		return
	}

	for _, parsedURL := range urls {
		normalizedLink, err := normalizeURL(parsedURL)
		if err != nil {
			fmt.Println("error normalizing link:", err)
			continue
		}

		cfg.mu.Lock()
		if len(cfg.pages) >= cfg.maxPages {
			cfg.mu.Unlock()
			continue
		}
		if _, seen := cfg.pages[normalizedLink]; seen {
			cfg.pages[normalizedLink]++
			cfg.mu.Unlock()
			continue
		} else {
			cfg.pages[normalizedLink] = 1
		}

		cfg.mu.Unlock()

		cfg.wg.Add(1)
		go cfg.crawlPage(parsedURL)
	}
}

func printReport(pages map[string]int, baseURL string) {
	fmt.Printf(`=============================
  REPORT for %s
=============================
`, baseURL)
	fmt.Println()
	pagesSlice := sortPages(pages)
	for _, page := range pagesSlice {

		fmt.Printf("Found %d internal links to %s\n", page.count, page.url)
	}
}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if _, visited := cfg.pages[normalizedURL]; visited {
		cfg.pages[normalizedURL]++
		return false
	}

	cfg.pages[normalizedURL] = 1
	return true
}

type Page struct {
	url   string
	count int
}

func sortPages(pages map[string]int) []Page {
	pagesSlice := make([]Page, 0, len(pages))
	for key, val := range pages {
		pagesSlice = append(pagesSlice, Page{
			url:   key,
			count: val,
		})
	}

	sort.Slice(pagesSlice, func(i, j int) bool {
		if pagesSlice[i] != pagesSlice[j] {
			return pagesSlice[i].count > pagesSlice[j].count
		}
		return pagesSlice[i].url < pagesSlice[j].url
	})

	return pagesSlice
}
