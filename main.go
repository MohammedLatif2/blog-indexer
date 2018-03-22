package main

import (
	"flag"

	"github.com/MohammedLatif2/blog-indexer/elastic"
	"github.com/MohammedLatif2/blog-indexer/elastic_driver"
	"github.com/MohammedLatif2/blog-indexer/http"
	"github.com/MohammedLatif2/blog-indexer/watcher"
)

func main() {
	var elURL, postsRoot string
	flag.StringVar(&postsRoot, "elURL", "/Users/malsayed/workspace/rayed.com/content/posts", "post directory")
	flag.StringVar(&elURL, "postsRoot", "http://localhost:9200/", "elastic host")
	flag.Parse()
	el := elastic.NewElastic(elURL)
	elm := elastic_driver.NewElasticDriver(el, postsRoot)
	go watcher.NewWatcher(postsRoot, elm.IndexDoc, elm.DeleteDoc).Start()
	s := http.NewServer(el)
	s.Start()
}
