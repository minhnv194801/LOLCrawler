package main

import (
	"lolcrawl/crawler"
)

func main() {
	crawler := crawler.NewLPLCrawler()
	defer crawler.Cancel()
	crawler.Start()
}
