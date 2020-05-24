package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/voldyman/emos"
	"github.com/voldyman/emos/internal/config"
	"golang.org/x/sync/errgroup"
)

const (
	workerCount = 10
)

var (
	searchFlag = flag.Bool("search", false, "searches an emoji")
	updateFlag = flag.Bool("update", false, "updates the emoji database")
)

func main() {
	flag.Parse()
	err := fmt.Errorf("no command specified")

	if *updateFlag {
		err = runUpdate()
	}

	if *searchFlag {
		err = runSearch(strings.Join(flag.Args(), " "))
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func runUpdate() error {
	emos, err := newEmos()
	if err != nil {
		return fmt.Errorf("failed to start emoji search: %w", err)
	}
	defer emos.Close()

	if err = emos.UpdateEmojis(); err != nil {
		return fmt.Errorf("failed to update emojis: %w", err)
	}
	fmt.Fprintln(os.Stderr, "updated local emojis")

	emos.RefreshIndex()
	fmt.Fprintln(os.Stderr, "updated emoji index")

	return nil
}

func newEmos() (*emos.EmojiSearch, error) {
	cfn := config.Loc(config.CacheFileName)
	ifn := config.Loc(config.IndexFileName)
	return emos.NewEmojiSearch(cfn, ifn)
}

type alfredResult struct {
	Items []*alfredItem `json:"items"`
}

type alfredItem struct {
	Title        string `json:"title"`
	QuickLookURL string `json:"quicklookurl"`
	Icon         struct {
		Type string `json:"type"`
		Path string `json:"path"`
	} `json:"icon"`
}

func newAlfredItem(title, path string) *alfredItem {
	ai := new(alfredItem)
	ai.Title = title
	ai.QuickLookURL = path
	ai.Icon.Type = "filepath"
	ai.Icon.Path = path

	return ai
}

func runSearch(input string) error {
	emos, err := newEmos()
	if err != nil {
		return fmt.Errorf("unable to start emoji search: %w", err)
	}
	defer emos.Close()

	result := emos.Search(input)

	aiChan := prepareResults(result.Emojis)

	alfredResult := alfredResult{Items: []*alfredItem{}}

	for item := range aiChan {
		alfredResult.Items = append(alfredResult.Items, item)
	}

	return json.NewEncoder(os.Stdout).Encode(alfredResult)
}

func prepareResults(emojis []*emos.Emoji) <-chan *alfredItem {
	aiChan := make(chan *alfredItem)
	g, _ := errgroup.WithContext(context.Background())

	workChan := make(chan *emos.Emoji)

	// worker pool
	for i := 0; i < workerCount; i++ {
		g.Go(func() error {
			for emoji := range workChan {

				imgPath, err := downloadImage(emoji)
				if err != nil {
					return err
				}
				name := emoji.Title
				aiChan <- newAlfredItem(name, imgPath)
			}
			return nil
		})
	}

	// send work to pool
	go func() {
		for _, emoji := range emojis {
			workChan <- emoji
		}
		close(workChan)
	}()

	// close pool when done
	go func() {
		err := g.Wait()
		if err != nil {
			fmt.Fprintf(os.Stderr, "fetching items failed: %+v\n", err)
		}
		close(aiChan)
	}()

	return aiChan
}

func downloadImage(emoji *emos.Emoji) (string, error) {
	cacheDir := getOrCreateCacheDir()
	emojiPath := filepath.Join(cacheDir, emoji.Title)
	emojiPath = fmt.Sprintf("%s.jpeg", emojiPath)

	if _, err := os.Stat(emojiPath); err == nil {
		return emojiPath, nil
	}

	resp, err := http.Get(emoji.Image)
	if err != nil {
		return "", fmt.Errorf("unable to download emoji image: %w", err)
	}
	defer resp.Body.Close()

	err = writeImage(resp.Body, emojiPath)
	if err != nil {
		return "", fmt.Errorf("unable to write image to disk: %w", err)
	}
	return emojiPath, nil
}

func getOrCreateCacheDir() string {
	imgCacheDir := config.Loc(config.ImageCacheDir)
	if _, err := os.Stat(imgCacheDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(imgCacheDir, 0755)
		}
	}

	return imgCacheDir
}

func writeImage(src io.Reader, path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("unable to open file for writing: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, src)
	if err != nil {
		return fmt.Errorf("unable to write response to file: %w", err)
	}
	return nil
}
