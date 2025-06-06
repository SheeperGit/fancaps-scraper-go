package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gocolly/colly"
	// "github.com/charmbracelet/bubbletea"
)

func getSearchQuery() string {
	fmt.Print("Enter Search Query: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return ""
}

func main() {
	var (
		query  = flag.String("q", "", "Search query term (required)")
		movies = flag.Bool("movies", true, "Include Movies in search query")
		tv     = flag.Bool("tv", true, "Include TV series in search query")
		anime  = flag.Bool("anime", true, "Include Anime in search query")
	)
	flag.Parse()

	if *query == "" {
		*query = getSearchQuery()
		if *query == "" {
			fmt.Fprintf(os.Stderr, "Error: Search query cannot be empty.\n")
			flag.Usage()
			os.Exit(1)
		}
	}

	params := url.Values{}
	params.Add("q", *query)
	if *movies {
		params.Add("MoviesCB", "Movies")
	}
	if *tv {
		params.Add("TVCB", "TV")
	}
	if *anime {
		params.Add("animeCB", "Anime")
	}
	params.Add("submit", "Submit Query")

	const baseURL = "https://fancaps.net/search.php"
	searchURL := baseURL + "?" + params.Encode()

	/* Create a Collector for FanCaps. */
	c := colly.NewCollector(
		colly.AllowedDomains("fancaps.net"),
		colly.Async(true),
	)

	/*
		On every h4 element which has an anchor child element,
		print the link text and the link itself.
	*/
	c.OnHTML("h4 > a", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
	})

	/* Before making a request, print "Visiting ..." */
	c.OnRequest(func(req *colly.Request) {
		fmt.Println("Visiting:", req.URL.String())
	})

	/* Start the collector. */
	c.Visit(searchURL)

	/* Wait for the collector to finish. (Required for Async) */
	c.Wait()
}
