package emos

import (
	"github.com/blevesearch/bleve"
	bleveMapping "github.com/blevesearch/bleve/mapping"
)

const maxBatchSize = 30

type index struct {
	idx bleve.Index
}

func NewMemIndex() (*index, error) {
	mapping := buildIndexMapping()

	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	return &index{idx}, nil
}
func NewIndex(loc string) (*index, error) {
	mapping := buildIndexMapping()
	idx, err := bleve.New(loc, mapping)
	if err != nil {
		return nil, err
	}
	return &index{
		idx: idx,
	}, nil
}

func OpenIndex(loc string) (*index, error) {
	idx, err := bleve.Open(loc)
	if err != nil {
		return nil, err
	}
	return &index{
		idx: idx,
	}, nil
}

func (i *index) Close() {
	i.idx.Close()
}

func (i *index) IndexEmoji(id string, e *Emoji) error {
	return i.idx.Index(id, e)
}
func (i *index) IndexEmojiStore(store map[string]*Emoji) error {
	batchCount := 0
	b := i.idx.NewBatch()
	for id, e := range store {
		b.Index(id, e)

		batchCount++
		if batchCount > maxBatchSize {
			err := i.idx.Batch(b)
			if err != nil {
				return err
			}

			b = i.idx.NewBatch()
			batchCount = 0
		}
	}
	return i.idx.Batch(b)
}

func (i *index) Delete(id string) error {
	return i.idx.Delete(id)
}

func (i *index) Search(text string) ([]string, error) {
	query := bleve.NewQueryStringQuery(text)
	search := bleve.NewSearchRequest(query)
	search.Size = 1000

	results, err := i.idx.Search(search)
	if err != nil {
		return []string{}, err
	}

	result := []string{}
	for _, h := range results.Hits {
		result = append(result, h.ID)
	}
	return result, nil
}

func (i *index) Count() int {
	c, err := i.idx.DocCount()
	if err != nil {
		return 0
	}
	// the world is 64 bit, amirite?
	return int(c)
}

func buildIndexMapping() bleveMapping.IndexMapping {
	emoteMapping := bleve.NewDocumentMapping()
	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("emoji", emoteMapping)

	fieldMapping := bleve.NewTextFieldMapping()
	fieldMapping.Store = false

	emoteMapping.AddFieldMappingsAt("Title", fieldMapping)
	emoteMapping.AddFieldMappingsAt("Slug", fieldMapping)
	emoteMapping.AddFieldMappingsAt("Description", fieldMapping)
	emoteMapping.AddFieldMappingsAt("Category", fieldMapping)

	return mapping
}
