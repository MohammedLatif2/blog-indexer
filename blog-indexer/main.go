package main

import (
	"log"
)

func check(e error) {
	if e != nil {
		log.Panicln("Error found ", e)
	}
}

func main() {
	el := NewElastic("http://localhost:9200/rayed/posts/")
	el.DeleteDoc("69f26ad3cbe76d607bfe4d73cb9eabadbfc8b573")
	StartWebServer()
}
