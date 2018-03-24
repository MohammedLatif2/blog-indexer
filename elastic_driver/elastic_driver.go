package elastic_driver

import (
	"crypto/sha1"
	"fmt"

	"github.com/MohammedLatif2/blog-indexer/config"
	"github.com/MohammedLatif2/blog-indexer/document"
	"github.com/MohammedLatif2/blog-indexer/elastic"
)

type ElasticDriver struct {
	El     *elastic.Elastic
	Config *config.Config
}

func NewElasticDriver(el *elastic.Elastic, config *config.Config) *ElasticDriver {
	return &ElasticDriver{El: el, Config: config}
}

func getIDX(filePath string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(filePath)))
}

func (elm *ElasticDriver) IndexDoc(filePath string) {
	id := getIDX(filePath)
	doc, _ := document.DocFromFile(filePath, elm.Config.HugoRoot)
	elm.El.IndexDoc(id, doc)
}

func (elm *ElasticDriver) DeleteDoc(filePath string) {
	id := getIDX(filePath)
	elm.El.DeleteDoc(id)
}
