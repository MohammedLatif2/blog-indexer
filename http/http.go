package http

import (
	"log"
	"net/http"

	"github.com/MohammedLatif2/blog-indexer/elastic"
)

type Server struct {
	el *elastic.Elastic
}

func NewServer(el *elastic.Elastic) *Server {
	return &Server{el}
}

func (server *Server) MakeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
	}
}

func (server *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	result, err := server.el.Search(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (server *Server) StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not implemented yet"))
}

func (server *Server) Start() {
	log.Println("Starting Web Server")
	http.HandleFunc("/search", server.MakeHandler(server.SearchHandler))
	http.HandleFunc("/stats", server.MakeHandler(server.StatsHandler))
	// http.HandleFunc("/delete", DeleteHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
