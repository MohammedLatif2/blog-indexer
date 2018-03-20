package elastic

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func NewElastic(baseUrl string) *Elastic {
	if !isEndedBySlash(baseUrl) {
		baseUrl = baseUrl + "/"
	}
	return &Elastic{BaseUrl: baseUrl}
}

func docFromFile(filePath, rootDirPath string) (*Document, error) {
	// Read file
	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	// Split header (frontmatter) and content (body)
	t := strings.Split(string(dat), "---")
	if len(t) < 3 {
		return nil, fmt.Errorf("Split arr is too small")
	}
	header := t[1]
	body := t[2]
	// Construct document
	h := Header{}
	yaml.Unmarshal([]byte(header), &h)
	date, err := time.Parse("2006-01-02T15:04:05-07:00", h.Date)
	if err != nil {
		return nil, err
	}
	d := Document{
		Path:       filePath[len(rootDirPath):],
		Text:       body,
		Title:      h.Title,
		Date:       date,
		Categories: h.Categories,
	}
	return &d, nil
}

func getIDX(filePath string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(filePath)))
}

func (el *Elastic) IndexDoc(filePath, rootDirPath string) error {
	idx := getIDX(filePath)
	doc, err := docFromFile(filePath, rootDirPath)
	if err != nil {
		return err
	}
	jsonDoc, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	// Index data
	url := el.BaseUrl + "rayed/post/" + idx
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonDoc))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return fmt.Errorf("document not indexed")
	}
	return nil
}

func (el *Elastic) DeleteDoc(filePath string) error {
	idx := getIDX(filePath)
	url := el.BaseUrl + "rayed/post/" + idx
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return fmt.Errorf("doc not deleted")
	}
	return nil
}

func (el *Elastic) Search(query string) ([]byte, error) {
	query = url.QueryEscape(query)
	reqURL := el.BaseUrl + "_search?q=" + query
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// parse json to result
	var result Result
	json.Unmarshal(bodyBytes, &result)
	docs := []Document{}
	// add docs from result
	for _, doc := range result.Hits.Hits {
		docs = append(docs, doc.Source)
	}
	// parse docs to json
	docsJson, err := json.Marshal(docs)
	if err != nil {
		return nil, err
	}
	return docsJson, nil
}

func isEndedBySlash(url string) bool {
	index := strings.LastIndexAny(url, "/")
	if index+1 == len(url) {
		return true
	}
	return false
}
