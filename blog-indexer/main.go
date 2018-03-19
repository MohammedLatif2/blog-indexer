package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	jww "github.com/spf13/jwalterweatherman"
	"gopkg.in/yaml.v2"
)

type Result struct {
	Hits Hits
}
type Hits struct {
	Total int
	Hits  []Hits2
}
type Hits2 struct {
	Source Document `json:"_source"`
}

type Header struct {
	Title      string
	Date       string
	Categories []string
}

type Document struct {
	Path       string
	Text       string
	Title      string
	Date       time.Time
	Categories []string
}

func check(e error) {
	if e != nil {
		log.Panicln("Error found ", e)
	}
}

func getFilesFrom(root string) []string {
	files := []string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() || strings.HasSuffix(info.Name(), ".md") == false {
			return nil
		}
		// fmt.Printf("visited file: %q\n", path)
		files = append(files, path)
		return nil
	})
	check(err)
	return files
}

func docFromFile(fileName, root string) *Document {
	// Read file
	dat, err := ioutil.ReadFile(fileName)
	check(err)
	// Split header (frontmatter) and content (body)
	t := strings.Split(string(dat), "---")
	if len(t) < 3 {
		log.Println("Split arr is too small")
		return nil
	}
	header := t[1]
	body := t[2]
	// Construct document
	h := Header{}
	yaml.Unmarshal([]byte(header), &h)
	date, err := time.Parse("2006-01-02T15:04:05-07:00", h.Date)
	check(err)
	d := Document{
		Path:       fileName[len(root):],
		Text:       body,
		Title:      h.Title,
		Date:       date,
		Categories: h.Categories,
	}
	return &d
}

func indexDoc(d *Document, idx string) {
	jsonData, err := json.Marshal(d)
	check(err)

	// Index data
	url := "http://localhost:9200/rayed/post/" + idx
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	check(err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()
}

func getIDX(file string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(file)))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	rootDir := "/Users/mabdullatif/Desktop/projects/rayed.com/content/posts/"
	files := getFilesFrom(rootDir)
	for _, file := range files {
		docIDX := getIDX(file)
		doc := docFromFile(file, rootDir)
		indexDoc(doc, docIDX)
	}
	w.Write([]byte("Done"))
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := url.QueryEscape(r.FormValue("q"))
	reqURL := "http://localhost:9200/rayed/post/_search?q=" + query
	resp, err := http.Get(reqURL)
	check(err)
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	check(err)
	var result Result
	json.Unmarshal(bodyBytes, &result)
	paths := []string{}
	for _, doc := range result.Hits.Hits {
		paths = append(paths, doc.Source.Path)
	}
	pathsJson, err := json.Marshal(paths)
	check(err)
	w.Write(pathsJson)
	// fmt.Println(result)
	// bodyString := string(bodyBytes)
	// w.Write([]byte(bodyString))
}

func main() {
	tweakLimit()
	dirs := getDirsFrom("/Users/mabdullatif/Desktop/projects/rayed.com/content/posts/")
	Watcher(dirs)
	http.HandleFunc("/index", indexHandler)
	http.HandleFunc("/search", searchHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func isDir(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return fi.Mode().IsDir()
}

func Watcher(dirs []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					if isDir(event.Name) {
						watcher.Add(event.Name)
						log.Println("Dir added to watchlist:", event.Name)
					}
				}

			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
	for _, dir := range dirs {
		err = watcher.Add(dir)
		if err != nil {
			log.Fatal(err)
		}
	}
	<-done
}

func getDirsFrom(root string) []string {
	dirs := []string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	check(err)
	return dirs
}

func tweakLimit() {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		jww.ERROR.Println("Unable to obtain rLimit", err)
	}
	if rLimit.Cur < rLimit.Max {
		rLimit.Max = 64000
		rLimit.Cur = 64000
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			jww.WARN.Println("Unable to increase number of open files limit", err)
		}
	}
}
