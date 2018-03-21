# Blog Indexer



## Building

    dep ensure
    go run main.go  # to run
    go build        # to build blog-indexer binary


## Testing with Elastic

You can use the project `docker-compose.yml` file to start an instance of Elasticsearch with Kibana useful for testing the app, or even tune it for production use:

    docker-compose up 

