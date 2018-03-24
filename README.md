
# Hugosearch

hugosearch is an application that is designed to index and search [hugo](https://gohugo.io/) websites, it is written in [Go](https://golang.org/)


## Building

    dep ensure
    go build    # to build blog-indexer binary

## Running

    cp config-sample.yml config.yml
    vi config.yml
    ./blog-indexer



## Docker 

### Building 

First we need to build the application with Linux as the target:

    GOOS=linux GOARCH=amd64 go build -o blog-indexer-linux

Next we need to build the Docker image:

    docker build -t hugosearch .


## Run Elastic Only in Docker

    docker-compose -f docker-compose-elastic-only.yml  up

### Run All Service

You can use the project `docker-compose.yml` file to start an instance of Elasticsearch with Kibana useful for testing the app, or even tune it for production use:

    docker-compose up 



## TODO

- Embed templates and static files inside binary
- Option to skip initial indexing
- Use environment variable for Dockerimage 