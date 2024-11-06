package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	//open text file with the urls that  I want to lookup for a rss feed
	file, err := os.Open("websites.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// Read text file line  by line
	scanner := bufio.NewScanner(file)
	// look if rss or atom feed exists
	rss_ext := [...]string{"rss.xml", "feed.xml", "feed.atom", "index.xml", "rss.atom", "rss.rss", "index.atom"}
	var feeds []string

	for scanner.Scan() {
		website_url := scanner.Text()
		for i := range len(rss_ext) {

			// Add / so it can properly be requested
			if len(website_url) > 0 && website_url[len(website_url)-1] != '/' {
				website_url = website_url + "/"

			}
			rss_url := website_url + rss_ext[i]
			resp, err := http.Get(rss_url)

			// Depending on the status code, we will append the url to our feeds
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Bad status: %d\n", resp.StatusCode)
			} else {
				feeds = append(feeds, rss_url)
			}
			if err != nil {
				fmt.Printf("Error fetching %s: %v\n", website_url, err)
				return
			}
			defer resp.Body.Close() // Important: always close the response body!
			fmt.Printf("Status Code: %d\n", resp.StatusCode)
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	for i := range len(feeds) {
		println(feeds[i])
	}
}
