package main

import (
	"bufio"
	"fmt"
	"log"
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
	for scanner.Scan() {

		website_url := scanner.Text()
		// look if rss or atom feed exists
		//NOTE: We are assuming that the rss feed will be located at website/rss.xml or website/atom.xml this might not be true for all of the websites
		rss_ext := [...]string{".rss", ".xml", ".atom"}
		// var feeds []string
		for i := range len(rss_ext) {
			rss_url := website_url + "rss" + rss_ext[i]
			// Add / so it can properly be requested
			if len(rss_url) > 0 && rss_url[len(rss_url)-1] != '/' {

				rss_url = rss_url + "/"

			}

			fmt.Println(rss_url)

			// resp, err := http.Get(website_url + rss_url)

			// Depending on the status code, we will append the url to our feeds
			// if resp.StatusCode != http.StatusOK {
			// 	fmt.Printf("Bad status: %d\n", resp.StatusCode)
			// } else {
			// 	feeds = append(feeds, rss_url)
			// 	fmt.Printf("adding", rss_url)
			//
			// }
			// if err != nil {
			// 	fmt.Printf("Error fetching %s: %v\n", website_url, err)
			// 	return
			// }
			// defer resp.Body.Close() // Important: always close the response body!
			// fmt.Printf("Status Code: %d\n", resp.StatusCode)
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
