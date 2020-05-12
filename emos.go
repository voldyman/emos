package emos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
	store map[string]*Emoji
	index *index
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
		store: store,
		index: idx,
	}, nil
}

func (f *EmojiSearch) IsIndexEmpty() bool {
	return f.index.Count() == 0
}

type SearchResult struct {
	Query  string
	Emojis []*Emoji
}

func (es *EmojiSearch) Search(input string) *SearchResult {
	r, err := es.index.Search(input)
	if err != nil {
		fmt.Println("unable to search", err)
		return nil
	}
	result := &SearchResult{
		Query:  input,
		Emojis: []*Emoji{},
	}

	for _, id := range r {
		if e, ok := es.store[id]; ok {
			result.Emojis = append(result.Emojis, e)
		}
	}
	return result
}

func (es *EmojiSearch) Close() {
	es.index.Close()
}

func (es *EmojiSearch) RefreshIndex() {
	es.index.IndexEmojiStore(es.store)
}

func getEmojis(cacheLoc string) (map[string]*Emoji, error) {
	_, err := os.Stat(cacheLoc)

	if err == nil {
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

	if os.IsNotExist(err) {
		emojis, err := fetchEmojis()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch emojis: %w", err)
		}

		d, err := json.Marshal(emojis)
		if err == nil {
			ioutil.WriteFile(cacheLoc, d, 0644)
		} else {
			fmt.Println("unable to cache emojis, ignoring")
		}
		return emojis, nil
	}

	return nil, fmt.Errorf("failed to stat cache file: %w", err)
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
