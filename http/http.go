package http

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"

	"github.com/MohammedLatif2/blog-indexer/elastic"
	log "github.com/Sirupsen/logrus"
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
	log.Infoln("index 1")
	q := r.FormValue("q")
	size := r.FormValue("size")
	from := r.FormValue("from")
	log.Infoln("index 2")

	result, err := server.el.Search(q, size, from)
	log.Infoln("index 3")
	// log.Println("Q:", q, "Result:", result)
	if err != nil {
		log.Warnln("IndexHandler: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
	log.Infoln("index 4")
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Warnln("IndexHandler: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
	t.Execute(w, map[string]interface{}{"q": q, "result": result})
}

func (server *Server) Panic(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("just_panic")
	t.Execute(w, nil)
}

func (server *Server) Start() {
	log.Infoln("Starting Web Server")

	r := mux.NewRouter()
	r.HandleFunc("/", server.IndexHandler)
	r.HandleFunc("/search", server.SearchHandler)
	r.HandleFunc("/stats", server.StatsHandler)
	r.HandleFunc("/panic", server.Panic)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.Handle("/", handlers.RecoveryHandler()(handlers.LoggingHandler(os.Stdout, r)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
