package http

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/MohammedLatif2/blog-indexer/elastic"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Server struct {
	el *elastic.Elastic
}

func NewServer(el *elastic.Elastic) *Server {
	return &Server{el}
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
	// log.Println("Q:", q, "Result:", result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, map[string]interface{}{"q": q, "result": result})
}

func (server *Server) JsHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/js.html")
	t.Execute(w, nil)
}

func (server *Server) Panic(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("just_panic")
	t.Execute(w, nil)
}

func (server *Server) Start() {
	log.Println("Starting Web Server")

	r := mux.NewRouter()
	r.HandleFunc("/", server.IndexHandler)
	r.HandleFunc("/js", server.JsHandler)
	r.HandleFunc("/search", server.SearchHandler)
	r.HandleFunc("/stats", server.StatsHandler)
	r.HandleFunc("/panic", server.Panic)

	http.Handle("/", handlers.RecoveryHandler()(handlers.LoggingHandler(os.Stdout, r)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
