package search

import (
	"sort"
)

// DocID is the use provided document ID
type DocID string

// internalDocID is our own representation of documents
type internalDocID uint64

// Searcher is used to find documents in the corpus
type Searcher struct {
	docIDMapping map[internalDocID]DocID

	trigramDocIDs map[ngram][]internalDocID
}

// Search finds documents which contain the provided text
func (s *Searcher) Search(text string) []DocID {
	ngrams := generateNgrams(text)

	foundNgrams := map[ngram]struct{}{}
	for _, n := range ngrams {
		if _, ok := s.trigramDocIDs[n]; !ok {
			continue
		}

		foundNgrams[n] = struct{}{}
	}

	if len(foundNgrams) == 0 {
		return []DocID{}
	}

	return s.convertInternalIDsToDocIds(s.findIntersectingDocIDs(convertSetToList(foundNgrams)))
}

func convertSetToList(set map[ngram]struct{}) []ngram {
	list := make([]ngram, 0, len(set))
	for k := range set {
		list = append(list, k)
	}
	return list
}

func (s *Searcher) findIntersectingDocIDs(found []ngram) []internalDocID {
	postingLists := make([][]internalDocID, 0, len(found))
	smallest := s.trigramDocIDs[found[0]]

	for _, f := range found {
		cur := s.trigramDocIDs[f]
		if len(cur) < len(smallest) {
			smallest = cur
			continue
		}
		postingLists = append(postingLists, cur)
	}

	freqCounter := newFreqCounter()

	for _, item := range smallest {

		for i := range postingLists {
			idx := sort.Search(len(postingLists[i]), func(idx int) bool {
				return postingLists[i][idx] >= item
			})

			if idx < len(postingLists[i]) && item == postingLists[i][idx] {
				freqCounter.Add(item)
			}
		}
	}
	return freqCounter.Sorted()
}

func (s *Searcher) convertInternalIDsToDocIds(ids []internalDocID) []DocID {
	result := make([]DocID, 0, len(ids))
	for _, id := range ids {
		result = append(result, s.docIDMapping[id])
	}
	return result
}

type freqCounter struct {
	counts map[internalDocID]int
}

func newFreqCounter() *freqCounter {
	return &freqCounter{
		counts: map[internalDocID]int{},
	}
}

func (fq *freqCounter) Add(id internalDocID) {
	if val, ok := fq.counts[id]; ok {
		fq.counts[id] = val + 1
		return
	}
	fq.counts[id] = 1
}

func (fq *freqCounter) Sorted() []internalDocID {
	arr := make([]freqPair, 0, len(fq.counts))

	for k, v := range fq.counts {
		arr = append(arr, freqPair{k, v})
	}
	sort.Sort(freqPairs(arr))

	result := make([]internalDocID, len(arr))
	for i, p := range arr {
		result[i] = p.doc
	}

	return result
}

type freqPair struct {
	doc  internalDocID
	freq int
}

type freqPairs []freqPair

func (f freqPairs) Len() int {
	return len([]freqPair(f))
}

func (f freqPairs) Less(i, j int) bool {
	return f[i].freq < f[j].freq
}

func (f freqPairs) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
