FROM alpine
COPY blog-indexer-linux run-hugosearch.sh /app/
COPY templates /app/templates
WORKDIR /app
EXPOSE 8080
CMD ["./run-hugosearch.sh"]
