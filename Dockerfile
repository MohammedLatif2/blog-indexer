FROM alpine
COPY blog-indexer /app/blog-indexer
WORKDIR /app
EXPOSE 8080
CMD ["./blog-indexer", "-postsRoot=/data", "-elURL=http://elasticsearch:9200"] 
