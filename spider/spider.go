package spider

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

// UserAgent is the default user agent of the Spider.
const UserAgent = "wospi - https://github.com/vlad-s/wospi"

// Options provides configuration params for the Spider.
type Options struct {
	MinLength int    // minimum length of words to be stored by the Spider
	MaxDepth  int    // maximum depth for the crawling
	UserAgent string // user agent used in HTTP requests

	// strict domain matching; if set, www sub-domain won't be added on non-www URL,
	// and non-www domain won't be added on www URL
	StrictDomain bool

	StripResult  bool
	StripCharset []rune
}

// Default applies default values to the Options.
func (o *Options) Default() {
	if o.UserAgent == "" {
		o.UserAgent = UserAgent
	}
}

// Spider provides the options and engines for the scraper.
type Spider struct {
	*Options

	collector *colly.Collector
	words     map[string]struct{}
	results   chan string
	done      chan struct{}
}

// Run starts the crawling & scraping on the provided URL.
func (s *Spider) Run(URL string) error {
	if URL == "" {
		return fmt.Errorf("URL can't be empty")
	}

	u, err := url.Parse(URL)
	if err != nil {
		return fmt.Errorf("couldn't parse URL: %s", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("URL must include scheme")
	}

	s.collector.AllowedDomains = []string{u.Host}

	if !s.StrictDomain {
		if strings.HasPrefix(u.Host, "www.") {
			// if domain contains www, add non-www domain to allowed domains
			s.collector.AllowedDomains = append(s.collector.AllowedDomains, strings.TrimLeft(u.Host, "www."))
		} else {
			// if domain does not contain www, add www domain to allowed domains
			s.collector.AllowedDomains = append(s.collector.AllowedDomains, "www."+u.Host)
		}
	}

	s.collector.OnHTML("a[href]", s.onHTML)
	s.collector.OnResponse(s.onResponse)

	err = s.collector.Visit(URL)
	if err != nil {
		return fmt.Errorf("couldn't start scraper: %s", err)
	}

	wg := &sync.WaitGroup{}
	go func(group *sync.WaitGroup) {
		wg.Add(1)
		for {
			select {
			case w := <-s.results:
				s.words[w] = struct{}{}
			case <-s.done:
				wg.Done()
				return
			}
		}
	}(wg)

	s.collector.Wait()
	s.done <- struct{}{}
	wg.Wait()

	for k := range s.words {
		fmt.Println(k)
	}

	return nil
}

func (s *Spider) onHTML(e *colly.HTMLElement) {
	e.Request.Visit(e.Attr("href"))
}

func stripResult(result string, charset []rune) string {
	for _, c := range charset {
		result = strings.Replace(result, string(c), "", -1)
	}
	return result
}

func (s *Spider) onResponse(r *colly.Response) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(doc.Find("body").Text()))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()

		if len(word) < s.MinLength {
			continue
		}

		if s.StripResult {
			stripped := stripResult(word, s.StripCharset)
			if len(stripped) < s.MinLength {
				continue
			}
			s.results <- stripped
		}

		s.results <- word
	}
}

func newCollector(opts *Options) *colly.Collector {
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(opts.MaxDepth),
		colly.UserAgent(opts.UserAgent),
	)

	c.DisableCookies()

	return c
}

// New returns a new Spider.
func New(opts *Options) *Spider {
	s := &Spider{
		Options: opts,

		collector: newCollector(opts),
		words:     make(map[string]struct{}, 0),
		results:   make(chan string),
		done:      make(chan struct{}),
	}

	return s
}
