package emos

import (
	"context"
	"fmt"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
)

const maxBatchSize = 30

type index struct {
	cfg bluge.Config
}

func NewMemIndex() (*index, error) {
	return &index{cfg: bluge.InMemoryOnlyConfig()}, nil
}

func NewIndex(loc string) (*index, error) {
	return &index{
		cfg: bluge.DefaultConfig(loc),
	}, nil
}

func OpenIndex(loc string) (*index, error) {
	return &index{
		cfg: bluge.DefaultConfig(loc),
	}, nil
}

func (i *index) Close() {
}

func (i *index) IndexEmoji(id string, e *Emoji) error {
	w, err := bluge.OpenWriter(i.cfg)
	if err != nil {
		return fmt.Errorf("unable to open writer: %w", err)
	}
	defer w.Close()

	err = w.Insert(createDocFromEmoji(id, e))

	if err != nil {
		return fmt.Errorf("unable to insert doc: %w", err)
	}
	return nil
}

func createDocFromEmoji(id string, e *Emoji) *bluge.Document {
	return bluge.NewDocument(id).
		AddField(bluge.NewTextField("Title", e.Title)).
		AddField(bluge.NewTextField("Description", e.Description)).
		AddField(bluge.NewTextField("Category", e.Category))
}

func (i *index) IndexEmojiStore(store map[string]*Emoji) error {
	batch := bluge.NewBatch()
	for id, e := range store {
		doc := createDocFromEmoji(id, e)
		batch.Update(doc.ID(), doc)
	}

	w, err := bluge.OpenWriter(i.cfg)
	if err != nil {
		return fmt.Errorf("unable to open writer: %w", err)
	}
	defer w.Close()

	err = w.Batch(batch)
	if err != nil {
		return fmt.Errorf("unable to write batch update to index; %w", err)
	}
	return nil
}

func (i *index) Delete(id string) error {
	w, err := bluge.OpenWriter(i.cfg)
	if err != nil {
		return fmt.Errorf("unable to open writer for delete: %w", err)
	}
	defer w.Close()

	doc := bluge.NewDocument(id)
	err = w.Delete(doc.ID())
	if err != nil {
		return fmt.Errorf("unable to delete item %s: %w", id, err)
	}
	return nil
}

func (i *index) Search(text string) (*searchIter, error) {
	r, err := bluge.OpenReader(i.cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to open index reader: %w", err)
	}

	titlePrefixQuery := bluge.NewPrefixQuery(text).SetField("Title")
	titleRegexQuery := bluge.NewRegexpQuery(".*" + text + ".*")

	categoryPrefixQuery := bluge.NewPrefixQuery(text).SetField("Category")
	categoryRegexQuery := bluge.NewRegexpQuery(text).SetField("Category")

	descFuzzyQuery := bluge.NewFuzzyQuery(text).SetField("Description")

	query := bluge.NewBooleanQuery().AddShould(
		titlePrefixQuery, titleRegexQuery,
		categoryPrefixQuery, categoryRegexQuery,
		descFuzzyQuery)

	req := bluge.NewAllMatches(query)

	iter, err := r.Search(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform search: %w", err)
	}

	return newSearchIter(iter, r), nil
}

func (i *index) Count() int {
	r, err := bluge.OpenReader(i.cfg)
	if err != nil {
		return 0
	}
	defer r.Close()

	count, err := r.Count()
	if err != nil {
		return 0
	}

	return int(count)
}

type searchIter struct {
	docIter   search.DocumentMatchIterator
	reader    *bluge.Reader
	lastError error
	match     *search.DocumentMatch
}

func newSearchIter(iter search.DocumentMatchIterator, r *bluge.Reader) *searchIter {
	return &searchIter{
		docIter:   iter,
		reader:    r,
		lastError: nil,
		match:     nil,
	}
}
func (s *searchIter) Next() (string, error) {
	defer func() {
		if s.lastError != nil && s.reader != nil {
			s.reader.Close()
			s.reader = nil
		}
	}()
	if s.lastError != nil {
		return "", s.lastError
	}

	s.match, s.lastError = s.docIter.Next()
	if s.lastError != nil {
		return "", s.lastError
	}

	if s.match == nil {
		s.lastError = fmt.Errorf("last match was nil")
		return "", s.lastError
	}
	id := ""
	s.lastError = s.match.VisitStoredFields(func(f string, value []byte) bool {
		if f == "_id" {
			id = string(value)
			return false
		}
		return true
	})
	return id, s.lastError
}
