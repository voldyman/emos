package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/voldyman/emos"
	"github.com/voldyman/emos/internal/config"
)

var (
	updateFlag   = flag.Bool("update", false, "update emoji list and index")
	markdownFlag = flag.Bool("md", false, "print markdown formatted link")
	onlyLinkFlag = flag.Bool("link", false, "only prints the link")
	luckyFlag    = flag.Bool("lucky", false, "only prints the first result")
	conigDirFlag = flag.Bool("cfg", false, "prints the config dir")
)

func init() {
	flag.Parse()
}

func main() {
	text := ""
	if flag.NArg() > 0 {
		text = strings.Join(flag.Args(), " ")
	}

	if *conigDirFlag {
		fmt.Println(config.Dir())
		return
	}

	e, err := emos.NewEmojiSearch(config.Loc(config.CacheFileName), config.Loc(config.IndexFileName))
	if err != nil {
		panic(err)
	}
	defer e.Close()

	if e.IsIndexEmpty() || *updateFlag {
		fmt.Println("building index, this will take a minute. you should hydrate. :blobsweat:")
		e.RefreshIndex()
	}

	if text == "" {
		// nothing to search, so let's not, eh?
		return
	}

	iter := e.Search(text)
	emoji, err := iter.Next()

	count := 20
	if *luckyFlag {
		count = 1
	}

	lines := []string{}
	for i := 0; i < count && err == nil; i++ {
		lines = append(lines, createPrintStatement(emoji))
		emoji, err = iter.Next()
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
