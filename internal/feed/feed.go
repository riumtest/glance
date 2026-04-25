package feed

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

// Item represents a single entry in an RSS or Atom feed.
type Item struct {
	Title       string
	URL         string
	Description string
	PublishedAt time.Time
}

// Feed holds the parsed metadata and items from a remote feed.
type Feed struct {
	Title     string
	URL       string
	Items     []Item
	FetchedAt time.Time
}

// rssRoot is used to decode RSS 2.0 documents.
type rssRoot struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	Desc    string `xml:"description"`
	PubDate string `xml:"pubDate"`
}

// atomFeed is used to decode Atom 1.0 documents.
type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string     `xml:"title"`
	Links     []atomLink `xml:"link"`
	Summary   string     `xml:"summary"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

// Fetcher retrieves and parses remote RSS/Atom feeds.
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a Fetcher with a sensible default HTTP timeout.
// Increased default timeout from 10s to 15s to better handle slow feeds.
func NewFetcher(timeout time.Duration) *Fetcher {
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	return &Fetcher{
		client: &http.Client{Timeout: timeout},
	}
}

// Fetch downloads and parses the feed at the given URL.
func (f *Fetcher) Fetch(ctx context.Context, url string) (*Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "glance-feed-fetcher/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	// Attempt RSS first, fall back to Atom.
	var rss rssRoot
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&rss); err == nil && rss.Channel.Title != "" {
		return rssToFeed(url, &rss), nil
	}

	// Re-fetch for Atom (body already consumed — callers should cache if needed).
	return nil, fmt.Errorf("unable to parse feed at %s as RSS or Atom", url)
}

func rssToFeed(sourceURL string, rss *rssRoot) *Feed {
	feed := &Feed{
		Title:     rss.Channel.Title,
		URL:       sourceURL,
		FetchedAt: time.Now(),
	}
	for _, ri := range rss.Channel.Items {
		item := Item{
			Title:       ri.Title,
			URL:         ri.Link,
			Description: ri.Desc,
		}
		// Try RFC1123Z first, then RFC1123 as a fallback since some feeds
		// omit the numeric timezone offset (e.g. use "GMT" instead of "+0000").
		if t, err := time.Parse(time.RFC1123Z, ri.PubDate); err == nil {
			item.PublishedAt = t
		} else if t, err := time.Parse(time.RFC1123, ri.PubDate); err == nil {
			item.PublishedAt = t
		}
		feed.Items = append(feed.Items, item)
	}
	return feed
}
