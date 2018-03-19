package elastic

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
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

type Elastic struct {
	BaseUrl string
}

func check(e error) {
	if e != nil {
		log.Panicln("Error found ", e)
	}
}

func NewElastic(baseUrl string) *Elastic {
	return &Elastic{BaseUrl: baseUrl}
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

func getIDX(file string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(file)))
}

func (el *Elastic) IndexDoc(file string, rootDir string) error {
	idx := getIDX(file)
	doc := docFromFile(file, rootDir)
	jsonData, err := json.Marshal(doc)
	check(err)

	// Index data
	url := el.BaseUrl + idx
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	check(err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()
	return nil
}

func (el *Elastic) DeleteDoc(id string) error {
	url := el.BaseUrl + id
	req, err := http.NewRequest("DELETE", url, nil)
	check(err)
	client := &http.Client{}
	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	check(err)
	fmt.Println(string(body))
	return nil
}

func (el *Elastic) Search(query string) ([]byte, error) {
	query = url.QueryEscape(query)
	reqURL := el.BaseUrl + "_search?q=" + query
	resp, err := http.Get(reqURL)
	check(err)
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	check(err)
	var result Result
	json.Unmarshal(bodyBytes, &result)
	docs := []Document{}
	for _, doc := range result.Hits.Hits {
		docs = append(docs, doc.Source)
	}
	docsJson, err := json.Marshal(docs)
	check(err)
	return docsJson, nil
}
