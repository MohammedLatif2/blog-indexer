package main

import (
	"log"

	"github.com/MohammedLatif2/blog-indexer/config"
	"github.com/MohammedLatif2/blog-indexer/elastic"
	"github.com/MohammedLatif2/blog-indexer/http"
	"github.com/MohammedLatif2/blog-indexer/watcher"
)

func main() {

	config, err := config.GetConfig()
	if err == nil {
		el := elastic.NewElastic(config.ElRoot)
		go watcher.Watcher(config.Root, el)
		s := http.NewServer(el)
		s.Start()
	}
	log.Println(err)
}
