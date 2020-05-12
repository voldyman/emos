package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/voldyman/emos"
)

func main() {
	text := ""
	if len(os.Args) > 1 {
		text = strings.Join(os.Args[1:], " ")
	}
	emos, err := emos.NewEmojiSearch("emojicache.json", "index.bleve")
	if err != nil {
		panic(err)
	}
	defer emos.Close()

	if emos.IsIndexEmpty() {
		fmt.Println("building index, this will take a minute. you should hydrate. :blobsweat:")
		emos.RefreshIndex()
	}

	result := emos.Search(text)

	for _, e := range result.Emojis {
		fmt.Printf("Title: %s, Image: %s\n", e.Title, e.Image)
	}
}
