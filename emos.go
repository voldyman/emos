package emos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/voldyman/emos/internal/config"
)

const (
	writePerms = 0644
)

type Emoji struct {
	Title       string
	Image       string
	Description string
	Category    string
	Width       int
	Height      int
}

type EmojiSearch struct {
	emojiCacheLoc string
	store         map[string]*Emoji
	index         *index
}

func NewEmojiSearch(cacheLoc, indexLoc string) (*EmojiSearch, error) {
	store, err := getEmojis(cacheLoc)
	if err != nil {
		return nil, fmt.Errorf("unable to create emoji search: %w", err)
	}

	idx, err := getIndex(indexLoc)
	if err != nil {
		return nil, fmt.Errorf("unable to create emoji search: %w", err)
	}

	return &EmojiSearch{
		emojiCacheLoc: cacheLoc,
		store:         store,
		index:         idx,
	}, nil
}

func (es *EmojiSearch) IsIndexEmpty() bool {
	return es.index.Count() == 0
}

type SearchResultIter struct {
	Query string
	es    *EmojiSearch
	iter  *searchIter
}

func (si *SearchResultIter) Next() (*Emoji, error) {
	docID, err := si.iter.Next()
	if err != nil {
		return nil, fmt.Errorf("stop searching: %w", err)
	}

	if emoji, ok := si.es.store[docID]; ok {
		return emoji, nil
	}

	return nil, fmt.Errorf("invalid state, docID: %s not found in store", docID)
}

func (es *EmojiSearch) Search(input string) *SearchResultIter {
	iter, err := es.index.Search(input)
	if err != nil {
		fmt.Println("unable to search", err)
		return nil
	}
	result := &SearchResultIter{
		Query: input,
		iter:  iter,
		es:    es,
	}
	return result
}

// Close closes the search
func (es *EmojiSearch) Close() {
	es.index.Close()
}

// RefreshIndex updates the index
func (es *EmojiSearch) RefreshIndex() {
	es.index.IndexEmojiStore(es.store)
}

// UpdateEmojis refreshes the local cache of emojis
func (es *EmojiSearch) UpdateEmojis() error {
	f, err := ioutil.TempFile("", "temp-emoji")
	if err != nil {
		return fmt.Errorf("unable to create temp file for caching emojis: %w", err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	_, err = updateEmojis(f.Name())
	if err != nil {
		return fmt.Errorf("unable to fetch new emojis: %w", err)
	}

	return os.Rename(f.Name(), config.Loc(config.CacheFileName))
}

func getEmojis(cacheLoc string) (map[string]*Emoji, error) {
	_, err := os.Stat(cacheLoc)

	if err == nil {
		if emojis, err := readEmojis(cacheLoc); err == nil {
			return emojis, nil
		}
	}

	return updateEmojis(cacheLoc)
}

func readEmojis(cacheLoc string) (map[string]*Emoji, error) {
	f, err := ioutil.ReadFile(cacheLoc)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	result := map[string]*Emoji{}
	err = json.Unmarshal(f, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cache: %w", err)
	}

	return result, nil

}

func updateEmojis(cacheLoc string) (map[string]*Emoji, error) {
	emojis, err := fetchEmojis()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch emojis: %w", err)
	}

	d, err := json.Marshal(emojis)
	if err == nil {
		ioutil.WriteFile(cacheLoc, d, writePerms)
	} else {
		fmt.Println("unable to cache emojis, ignoring")
	}

	return emojis, nil
}

func getIndex(indexLoc string) (*index, error) {
	_, err := os.Stat(indexLoc)

	if err == nil {
		idx, err := OpenIndex(indexLoc)
		if err != nil {
			return nil, fmt.Errorf("failed to open index: %w", err)
		}
		return idx, nil
	}

	if os.IsNotExist(err) {
		idx, err := NewIndex(indexLoc)
		if err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
		return idx, nil
	}

	return nil, fmt.Errorf("failed to stat index: %w", err)
}
