package elastic_manager

import (
	"crypto/sha1"
	"fmt"

	"github.com/MohammedLatif2/blog-indexer/document_manager"
	"github.com/MohammedLatif2/blog-indexer/elastic"
)

type ElasticManager struct {
	El *elastic.Elastic
}

func newElasticManager(el *elastic.Elastic) *ElasticManager {
	return &ElasticManager{El: el}
}

func getIDX(filePath string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(filePath)))
}

func (elm *ElasticManager) IndexDocManager(filePath, rootDirPath string) error {
	id := getIDX(filePath)
	doc, err := document_manager.DocFromFile(filePath, rootDirPath)
	if err != nil {
		return err
	}
	elm.El.IndexDoc(id, doc)
	return nil
}

func (elm *ElasticManager) DeleteDocManager(filePath string) {
	id := getIDX(filePath)
	elm.El.DeleteDoc(id)
}
