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
	_Idx       string
	Path       string
	Text       string
	Title      string
	Date       time.Time
	Categories []string
}
type Job struct {
	Command  string
	Id       string
	Document *Document
}
type Index struct {
	Index *Index1 `json:"index"`
}
type Index1 struct {
	Index string `json:"_index"`
	Type  string `json:"_type"`
	Id    string `json:"_id"`
}

type Elastic struct {
	BaseUrl string
	jobs    chan *Job
	done    chan struct{}
}

func NewElastic(baseUrl string) *Elastic {
	if !isEndedBySlash(baseUrl) {
		baseUrl = baseUrl + "/"
	}
	el := &Elastic{}
	el.BaseUrl = baseUrl
	el.jobs = make(chan *Job, 1)
	el.done = make(chan struct{}, 1)

	go el.batcher()

	return el
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
	el.jobs <- &Job{
		Command:  "index",
		Document: doc,
		Id:       idx,
	}
	return nil
}

func (el *Elastic) batcher() {
	tick := time.Tick(1 * time.Second)
	jobs := make([]*Job, 0)
OuterLoop:
	for {
		select {
		case job := <-el.jobs:
			jobs = append(jobs, job)
			if len(jobs) == 100 {
				el.bulkJob(jobs)
				jobs = make([]*Job, 0)
			}
		case <-tick:
			if len(jobs) == 0 {
				continue
			}
			el.bulkJob(jobs)
			jobs = make([]*Job, 0)
		case <-el.done:
			el.bulkJob(jobs)
			break OuterLoop
		}
	}
}

func (el *Elastic) Close() {
	el.done <- struct{}{}
}

func (el *Elastic) bulkJob(jobs []*Job) {
	url := el.BaseUrl + "_bulk"

	lines := []string{}
	for _, job := range jobs {
		command := map[string]interface{}{
			job.Command: map[string]string{
				"_index": "rayed",
				"_type":  "post",
				"_id":    job.Id,
			},
		}
		commandJSON, err := json.Marshal(command)
		if err != nil {
			log.Println(err.Error())
		}
		lines = append(lines, string(commandJSON))

		if job.Document != nil {
			docJSON, err := json.Marshal(job.Document)
			if err != nil {
				log.Println(err.Error())
			}
			lines = append(lines, string(docJSON))
		}
	}

	request := strings.Join(lines, "\n") + "\n"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(request)))
	req.Header.Set("Content-Type", "application/x-ndjson")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.Println("document not indexed")
	}
}

func (el *Elastic) DeleteDoc(filePath string) error {
	idx := getIDX(filePath)
	el.jobs <- &Job{
		Command: "delete",
		Id:      idx,
	}
	return nil
}

func (el *Elastic) Search(query string, size string, from string) ([]byte, error) {
	query = url.QueryEscape(query)
	if len(query) == 0 {
		return nil, fmt.Errorf("query param is empty")
	}
	reqURL := el.BaseUrl + "_search?q=" + query
	if len(size) != 0 {
		reqURL = reqURL + "&size=" + size
	}
	if len(from) != 0 {
		reqURL = reqURL + "&from=" + from
	}
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
