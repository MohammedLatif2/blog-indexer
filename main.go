package main

import (
	"flag"
	"log"
	"time"

	"github.com/MohammedLatif2/blog-indexer/config"
	"github.com/MohammedLatif2/blog-indexer/elastic"
	"github.com/MohammedLatif2/blog-indexer/elastic_driver"
	"github.com/MohammedLatif2/blog-indexer/http"
	"github.com/MohammedLatif2/blog-indexer/watcher"
)

func main() {
	configFile := ""
	flag.StringVar(&configFile, "config", "config.yml", "Configuration file")
	flag.Parse()

	// Read configuration
	config, err := config.NewConfig(configFile)
	if err != nil {
		log.Fatal("Couldn't open config file")
	}
	log.Println(config)

	// Wait for Elastic to be ready!
	el := elastic.NewElastic(config)
	for {
		if err := el.Ready(); err != nil {
			log.Println("Elastic not ready:", err)
			time.Sleep(time.Second)
			continue
		} else {
			break
		}
	}

	elm := elastic_driver.NewElasticDriver(el, config)
	go watcher.NewWatcher(config.Hugo.ContentRoot, elm.IndexDoc, elm.DeleteDoc).Start()

	s := http.NewServer(el)
	s.Start()
}
