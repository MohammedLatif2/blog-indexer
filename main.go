package main

import (
	"github.com/MohammedLatif2/blog-indexer/elastic"
	"github.com/MohammedLatif2/blog-indexer/http"
	"github.com/MohammedLatif2/blog-indexer/watcher"
)

func main() {
	root := "/Users/mabdullatif/Desktop/projects/rayed.com/content/posts"
	elRoot := "http://localhost:9200/rayed/posts/"
	el := elastic.NewElastic(elRoot)
	go watcher.Watcher(root, el)
	http.StartWebServer()
}
