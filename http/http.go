package http

import (
	"encoding/json"
	"html/template"
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
	size := r.FormValue("size")
	from := r.FormValue("from")
	result, err := server.el.Search(query, size, from)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
	docsJson, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(docsJson)
}

func (server *Server) StatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not implemented yet"))
}

func (server *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")
	size := r.FormValue("size")
	from := r.FormValue("from")

	result, err := server.el.Search(q, size, from)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, map[string]interface{}{"q": q, "result": result})
}

func (server *Server) Start() {
	log.Println("Starting Web Server")
	http.HandleFunc("/search", server.MakeHandler(server.SearchHandler))
	http.HandleFunc("/stats", server.MakeHandler(server.StatsHandler))
	http.HandleFunc("/", server.MakeHandler(server.IndexHandler))
	// http.HandleFunc("/delete", DeleteHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
