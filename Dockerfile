FROM alpine
COPY blog-indexer-linux run-hugosearch.sh /app/
COPY templates /app/templates
COPY static /app/static
WORKDIR /app
EXPOSE 8080
CMD ["./run-hugosearch.sh"]
