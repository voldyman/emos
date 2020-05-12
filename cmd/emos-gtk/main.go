package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/voldyman/emos"
)

type searcher struct {
	sync.Mutex
	s *emos.EmojiSearch
}

func NewSearcher() *searcher {
	s := &searcher{}
	return s
}

func (s *searcher) FixUp() {
	go func() {
		es, err := emos.NewEmojiSearch("cache.json", "emoji.emosi")
		if err != nil {
			panic(err)
		}

		if es.IsIndexEmpty() {
			es.RefreshIndex()
		}

		s.Lock()
		s.s = es
		s.Unlock()
	}()
}
func (s *searcher) Search(input string) *emos.SearchResult {
	s.Lock()
	defer s.Unlock()
	if s.s == nil {
		return &emos.SearchResult{
			Query:  input,
			Emojis: []*emos.Emoji{},
		}
	}
	return s.s.Search(input)
}

func main() {
	gtk.Init(&os.Args)

	app, err := gtk.ApplicationNew("net.tripent.emos", glib.APPLICATION_NON_UNIQUE)
	if err != nil {
		log.Fatal("couldn't create application", err)
	}

	es := NewSearcher()
	app.Connect("activate", func() {
		onActivate(es, app)
		es.FixUp()
	})

	os.Exit(app.Run(os.Args))
}

func onActivate(es *searcher, app *gtk.Application) {
	curWin := app.GetActiveWindow()
	if curWin != nil {
		curWin.Present()
		return
	}
	win, err := gtk.ApplicationWindowNew(app)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("Emoji Search")
	win.SetDefaultSize(800, 600)
	win.Connect("destroy", func() {
		app.Quit()
	})

	container, err := gtk.GridNew()
	if err != nil {
		log.Fatal("unable to create grid")
	}

	container.SetVExpand(true)
	container.SetHExpand(true)
	win.Add(container)

	searchEntry, err := gtk.SearchEntryNew()
	if err != nil {
		log.Fatal("unabel to create search entry")
	}

	searchEntry.SetHExpand(true)
	searchEntry.SetMarginStart(10)
	searchEntry.SetMarginEnd(10)
	searchEntry.SetMarginTop(5)
	searchEntry.SetMarginBottom(5)

	container.Attach(searchEntry, 0, 0, 1, 1)

	listBox, err := gtk.ListBoxNew()
	if err != nil {
		log.Fatal("unable to create list box", err)
	}

	listBox.SetVExpand(true)
	listBox.SetHExpand(true)
	container.Attach(listBox, 0, 1, 1, 1)

	searchEntry.Connect("search-changed", func() {
		stext, _ := searchEntry.GetText()
		result := es.Search(stext)
		fmt.Println("Searched:", stext, ", Results:", len(result.Emojis))

		listBox.GetChildren().Foreach(func(item interface{}) {
			listBox.Remove(item.(*gtk.Widget))
		})
		for _, r := range result.Emojis {
			r, err := createListRow(r)
			if err != nil {
				log.Fatal(err)
			}
			listBox.Add(r)
		}
		listBox.ShowAll()
	})

	app.AddWindow(win)
	win.ShowAll()
	searchEntry.GrabFocus()
}

func createListRow(e *emos.Emoji) (*gtk.ListBoxRow, error) {
	row, err := gtk.ListBoxRowNew()
	if err != nil {
		return nil, err
	}
	row.SetMarginTop(2)
	row.SetMarginStart(5)
	row.SetMarginEnd(5)

	img, err := gtk.ImageNewFromFile(e.Image)
	if err != nil {
		return nil, err
	}

	label, err := gtk.LabelNew(e.Title)
	if err != nil {
		return nil, err
	}
	label.SetMarginTop(2)
	label.SetMarginBottom(2)
	label.SetHAlign(gtk.ALIGN_START)

	row.Add(img)
	row.Add(label)

	return row, nil
}
