package main

import (
	"fmt"
	"net/url"
	"os"
	"time"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	}

	if len(args) > 1 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}

	fmt.Printf("\n ==starting crawl of: %s ==\n", args[0])
	url, err := url.Parse(args[0])
	if err != nil {
		fmt.Printf("could not parse URL: %v", err)
		os.Exit(1)
	}

	cfg := NewConfig(10, url)
	crawl(cfg, *url)
	for k, v := range cfg.pages {
		fmt.Printf("site: %s count: %d \n", k, v)
	}
	fmt.Print("\n ==Finished== \n")
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

//1 Thread Crawling took 7.562786s
//2 Threads Crawling took 3.596488834s
//5 Threads Crawling took 1.825310417s
//10 Threads Crawling took 1.014717666s
