package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/voldyman/emos"
)

var (
	updateFlag   = flag.Bool("update", false, "update emoji list and index")
	markdownFlag = flag.Bool("md", false, "print markdown formatted link")
	onlyLinkFlag = flag.Bool("link", false, "only prints the link")
	luckyFlag    = flag.Bool("lucky", false, "only prints the first result")
)

func init() {
	flag.Parse()
}

func main() {
	text := ""
	if len(flag.Args()) > 0 {
		text = strings.Join(os.Args[1:], " ")
	}
	emos, err := emos.NewEmojiSearch(configFile("emojicache.json"), configFile("index.bleve"))
	if err != nil {
		panic(err)
	}
	defer emos.Close()

	if emos.IsIndexEmpty() || *updateFlag {
		fmt.Println("building index, this will take a minute. you should hydrate. :blobsweat:")
		emos.RefreshIndex()
	}

	result := emos.Search(text)
	emojis := result.Emojis

	if *luckyFlag && len(emojis) > 1 {
		emojis = emojis[0:1]
	}

	lines := []string{}
	for _, e := range emojis {
		lines = append(lines, createPrintStatement(e))
	}

	if isStdoutPiped() {
		fmt.Printf("%s", strings.Join(lines, "\n"))
	} else {
		fmt.Println(strings.Join(lines, "\n"))
	}
}

func createPrintStatement(e *emos.Emoji) string {
	var b strings.Builder
	if !*onlyLinkFlag {
		b.WriteString(e.Title)
		b.WriteString(" - ")
	}
	if *markdownFlag {
		b.WriteString("/md ![](")
		b.WriteString(e.Image)
		b.WriteString(")")
	} else {
		b.WriteString(e.Image)
	}
	return b.String()
}

func isStdoutPiped() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice == 0
}

func configFile(name string) string {
	return filepath.Join(emosDir(), name)
}

func emosDir() string {
	cfg, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("failed to get config dir")
		panic(err)
	}

	return filepath.Join(cfg, "emos")
}
