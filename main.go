package main

import (
	"log"

	"github.com/MohammedLatif2/blog-indexer/elastic"
	"github.com/MohammedLatif2/blog-indexer/http"
	"github.com/MohammedLatif2/blog-indexer/watcher"
)

func check(e error) {
	if e != nil {
		log.Panicln("Error found ", e)
	}
}

func main() {
	root := "/Users/mabdullatif/Desktop/projects/rayed.com/content/posts"
	go watcher.Watcher(root)
	el := elastic.NewElastic("http://localhost:9200/rayed/posts/")
	el.DeleteDoc("69f26ad3cbe76d607bfe4d73cb9eabadbfc8b573")
	http.StartWebServer()
}
