package document_manager

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

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

func DocFromFile(filePath, rootDirPath string) (*Document, error) {
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
