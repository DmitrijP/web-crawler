package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	}

	if len(args) > 3 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}

	url, err := url.Parse(args[0])
	if err != nil {
		fmt.Printf("could not parse URL: %v", err)
		os.Exit(1)
	}

	threads := 1
	maxPages := 10

	if len(args) > 1 {
		t, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Printf("could not parse thread arg: %v", err)
		} else {
			threads = t
		}
	}
	if len(args) > 2 {
		p, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Printf("could not parse thread arg: %v", err)
		} else {
			maxPages = p
		}
	}

	fmt.Printf("\n ==starting crawl of: %s threads:%d maxPages:%d ==\n", args[0], threads, maxPages)
	cfg := NewConfig(threads, maxPages, url)
	crawl(cfg, *url)
	fmt.Print("\n ==Finished== \n")
	printReport(cfg.pages, url.String())
}

func crawl(cfg *config, url url.URL) {
	defer trackTime(time.Now(), "Crawling")
	cfg.wg.Add(1)
	go cfg.crawlPage(url.String())
	cfg.wg.Wait()
}

func trackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}

func printReport(pages map[string]int, baseURL string) {
	fmt.Printf("=============================\nREPORT for %s\n=============================\n", baseURL)
	for k, v := range pages {
		fmt.Printf("Found %d internal links to %s\n", v, k)
	}
}

//1 Thread Crawling took 7.562786s
//2 Threads Crawling took 3.596488834s
//5 Threads Crawling took 1.825310417s
//10 Threads Crawling took 1.014717666s
