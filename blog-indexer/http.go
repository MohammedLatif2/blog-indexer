package main

import (
	"log"
	"net/http"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	el := NewElastic("http://localhost:9200/rayed/posts/")
	query := r.FormValue("q")
	result, _ := el.Search(query)
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not implemented yet"))
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	// query := r.FormValue("q")
	// result := DeleteDoc(query)
}

func StartWebServer() {
	log.Println("Starting Web Server")
	http.HandleFunc("/search", SearchHandler)
	http.HandleFunc("/stats", StatsHandler)
	// http.HandleFunc("/delete", DeleteHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
