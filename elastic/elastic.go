package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MohammedLatif2/blog-indexer/config"
	"github.com/MohammedLatif2/blog-indexer/document"
)

type Result struct {
	Hits Hits
}
type Hits struct {
	Total int
	Hits  []Hits2
}
type Hits2 struct {
	Source document.Document `json:"_source"`
}

type Job struct {
	Command  string
	Id       string
	Document *interface{}
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
	Config *config.Config
	jobs   chan *Job
	done   chan struct{}
}

func NewElastic(config *config.Config) *Elastic {
	el := &Elastic{}
	el.Config = config
	el.jobs = make(chan *Job, 1)
	el.done = make(chan struct{}, 1)

	go el.batcher()

	return el
}

func (el *Elastic) IndexDoc(id string, doc interface{}) {
	el.jobs <- &Job{
		Command:  "index",
		Document: &doc,
		Id:       id,
	}
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
	url := fmt.Sprintf("%s/_bulk", el.Config.Elastic.Base)

	lines := []string{}
	for _, job := range jobs {
		command := map[string]interface{}{
			job.Command: map[string]string{
				"_index": el.Config.Elastic.Index,
				"_type":  el.Config.Elastic.Type,
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

func (el *Elastic) DeleteDoc(id string) {
	el.jobs <- &Job{
		Command: "delete",
		Id:      id,
	}
}

func (el *Elastic) Search(query string, size string, from string) ([]document.Document, error) {
	query = url.QueryEscape(query)
	if len(query) == 0 {
		return nil, nil
	}
	reqURL := fmt.Sprintf("%s/%s/%s/_search?q=%s", el.Config.Elastic.Base, el.Config.Elastic.Index, el.Config.Elastic.Type, query)
	if len(size) != 0 {
		reqURL = reqURL + "&size=" + size
	}
	if len(from) != 0 {
		reqURL = reqURL + "&from=" + from
	}
	log.Println(reqURL)
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
	docs := []document.Document{}
	// add docs from result
	for _, doc := range result.Hits.Hits {
		docs = append(docs, doc.Source)
	}
	return docs, nil
	// parse docs to json

	// return docsJson, nil
}
