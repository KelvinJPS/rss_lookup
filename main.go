package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Basic RSS/Atom feed structures for validation
type RSS struct {
	XMLName xml.Name    `xml:"rss"`
	Channel *RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title string `xml:"title"`
}

type Atom struct {
	XMLName xml.Name `xml:"feed"`
	Title   string   `xml:"title"`
}

// OPML structures
type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    OPMLHead `xml:"head"`
	Body    OPMLBody `xml:"body"`
}

type OPMLHead struct {
	Title        string `xml:"title"`
	DateCreated  string `xml:"dateCreated"`
	DateModified string `xml:"dateModified"`
	OwnerName    string `xml:"ownerName"`
	OwnerEmail   string `xml:"ownerEmail"`
}

type OPMLBody struct {
	Outlines []OPMLOutline `xml:"outline"`
}

type OPMLOutline struct {
	Type    string `xml:"type,attr"`
	Text    string `xml:"text,attr"`
	Title   string `xml:"title,attr,omitempty"`
	XMLURL  string `xml:"xmlUrl,attr"`
	HTMLURL string `xml:"htmlUrl,attr,omitempty"`
}

type FeedInfo struct {
	URL   string
	Title string
}

func getTitleFromFeed(body []byte) string {
	// Try parsing as RSS
	var rss RSS
	if err := xml.Unmarshal(body, &rss); err == nil && rss.Channel != nil {
		return rss.Channel.Title
	}

	// Try parsing as Atom
	var atom Atom
	if err := xml.Unmarshal(body, &atom); err == nil {
		return atom.Title
	}

	return ""
}

func isValidFeed(resp *http.Response) (bool, string) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, ""
	}

	// Try parsing as RSS
	var rss RSS
	if err := xml.Unmarshal(body, &rss); err == nil && rss.XMLName.Local == "rss" {
		return true, getTitleFromFeed(body)
	}

	// Try parsing as Atom
	var atom Atom
	if err := xml.Unmarshal(body, &atom); err == nil && atom.XMLName.Local == "feed" {
		return true, getTitleFromFeed(body)
	}

	return false, ""
}

func writeOPMLFile(feeds []FeedInfo, filename string) error {
	opml := OPML{
		Version: "2.0",
		Head: OPMLHead{
			Title:        "RSS Feeds",
			DateCreated:  time.Now().Format(time.RFC1123Z),
			DateModified: time.Now().Format(time.RFC1123Z),
		},
		Body: OPMLBody{
			Outlines: make([]OPMLOutline, 0, len(feeds)),
		},
	}

	for _, feed := range feeds {
		outline := OPMLOutline{
			Type:    "rss",
			Text:    feed.Title,
			Title:   feed.Title,
			XMLURL:  feed.URL,
			HTMLURL: strings.TrimSuffix(feed.URL[:strings.LastIndex(feed.URL, "/")], "/"),
		}
		opml.Body.Outlines = append(opml.Body.Outlines, outline)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")

	// Write XML header
	file.WriteString(xml.Header)

	return encoder.Encode(opml)
}

func main() {
	file, err := os.Open("websites.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	rssExtensions := []string{
		"rss.xml",
		"feed.xml",
		"feed.atom",
		"index.xml",
		"rss.atom",
		"rss",
		"feed",
		"index.atom",
	}

	var validFeeds []FeedInfo
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	for scanner.Scan() {
		websiteURL := strings.TrimSpace(scanner.Text())
		if websiteURL == "" {
			continue
		}

		if !strings.HasPrefix(websiteURL, "http://") && !strings.HasPrefix(websiteURL, "https://") {
			websiteURL = "https://" + websiteURL
		}

		if !strings.HasSuffix(websiteURL, "/") {
			websiteURL += "/"
		}

		for _, ext := range rssExtensions {
			rssURL := websiteURL + ext
			resp, err := client.Get(rssURL)
			if err != nil {
				fmt.Printf("Error fetching %s: %v\n", rssURL, err)
				continue
			}

			if resp.StatusCode == http.StatusOK {
				if isValid, title := isValidFeed(resp); isValid {
					feedInfo := FeedInfo{
						URL:   rssURL,
						Title: title,
					}
					if feedInfo.Title == "" {
						feedInfo.Title = websiteURL // fallback to URL if no title found
					}
					validFeeds = append(validFeeds, feedInfo)
					fmt.Printf("Found valid feed: %s (%s)\n", rssURL, title)
				} else {
					fmt.Printf("Invalid feed format: %s\n", rssURL)
				}
			} else {
				fmt.Printf("Bad status for %s: %d\n", rssURL, resp.StatusCode)
			}
			resp.Body.Close()
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Write to OPML file
	if err := writeOPMLFile(validFeeds, "feeds.opml"); err != nil {
		log.Fatal("Error writing OPML file:", err)
	}

	fmt.Printf("\nFound %d valid feeds, written to feeds.opml\n", len(validFeeds))
}

