package emos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func fetchEmojis() (map[string]*Emoji, error) {
	categories, err := fetchEmojiCategories()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch emoji categories: %w", err)
	}

	apiEmojis, err := fetchRawEmojis()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch raw emojis: %w", err)
	}

	result := map[string]*Emoji{}

	for _, e := range apiEmojis {
		category := strconv.Itoa(e.Category)
		if name, ok := categories[category]; ok {
			category = name
		}

		result[strconv.Itoa(e.ID)] = &Emoji{
			Title:       e.Title,
			Image:       e.Image,
			Description: e.Description,
			Category:    category,
			Width:       e.Width,
			Height:      e.Height,
		}
	}
	return result, nil
}

type apiEmoji struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Category    int    `json:"category"`
	License     string `json:"license"`
	Faves       int    `json:"faves"`
	SubmittedBy string `json:"submitted_by"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	FileSize    int    `json:"filesize"`
}

func fetchRawEmojis() ([]apiEmoji, error) {
	resp, err := http.Get("https://emoji.gg/api/")
	if err != nil {
		return nil, err
	}

	jData, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var result []apiEmoji
	err = json.Unmarshal(jData, &result)

	return result, err
}

func fetchEmojiCategories() (map[string]string, error) {
	resp, err := http.Get("https://emoji.gg/api/?request=categories")
	if err != nil {
		return nil, err
	}

	jData, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var result map[string]string
	err = json.Unmarshal(jData, &result)

	return result, err

}
