package search

import (
	"sort"
)

// Indexable is responsible for providing text to index from an object
type Indexable interface {
	IndexText() string
}

// Builder used to build a searchable
type Builder struct {
	idMapping  map[internalDocID]DocID
	lastID     internalDocID
	trigramIdx map[ngram][]internalDocID
}

// NewBuilder creates a new builder
func NewBuilder() *Builder {
	return &Builder{
		idMapping:  map[internalDocID]DocID{},
		lastID:     0,
		trigramIdx: map[ngram][]internalDocID{},
	}
}

// AddDoc adds a document to builder
func (b *Builder) AddDoc(id DocID, d Indexable) {
	iid := b.nextInternalID(id)
	grams := generateNgrams(d.IndexText())

	for _, g := range grams {
		b.addToIndex(iid, g)
	}
}

// Searcher creates a searcher from builder
func (b *Builder) Searcher() *Searcher {
	return &Searcher{
		docIDMapping:  b.idMapping,
		trigramDocIDs: b.trigramIdx,
	}
}

func (b *Builder) nextInternalID(id DocID) internalDocID {
	iid := b.lastID
	b.lastID++
	b.idMapping[iid] = id
	return iid
}

func (b *Builder) addToIndex(id internalDocID, gram ngram) {
	if ids, ok := b.trigramIdx[gram]; ok {
		idx := sort.Search(len(ids), func(i int) bool { return ids[i] >= id })
		if idx < len(ids) && ids[idx] == id {
			return
		}
		ids = append(ids, id)
		b.trigramIdx[gram] = ids
		return
	}

	b.trigramIdx[gram] = []internalDocID{id}
}
