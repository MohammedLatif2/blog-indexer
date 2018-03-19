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
	StartWebServer()
}
