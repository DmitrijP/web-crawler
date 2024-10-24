package main

import (
	"fmt"
	"os"
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
	pages := make(map[string]int, 0)
	crawlPage(args[0], args[0], pages)
	for k, v := range pages {
		fmt.Printf("site: %s count: %d \n", k, v)
	}
	fmt.Print("\n ==Finished== \n")
}
