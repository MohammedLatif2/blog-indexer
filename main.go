package main

import (
	"flag"

	"github.com/MohammedLatif2/blog-indexer/elastic"
	"github.com/MohammedLatif2/blog-indexer/http"
)

func main() {
	var elURL, postsRoot string
	flag.StringVar(&postsRoot, "elURL", "/Users/malsayed/workspace/rayed.com/content/posts", "post directory")
	flag.StringVar(&elURL, "postsRoot", "http://localhost:9200/", "elastic host")
	flag.Parse()
	el := elastic.NewElastic(elURL)
	// go watcher.NewWatcher(postsRoot, el.IndexDoc, el.DeleteDoc).Start()
	s := http.NewServer(el)
	s.Start()
}
