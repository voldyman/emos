package search

import (
	"testing"
)

type searchable struct {
	text string
}

func (s searchable) IndexText() string {
	return s.text
}

func TestSearch(t *testing.T) {
	b := NewBuilder()
	b.AddDoc("a", searchable{text: "sad"})
	s := b.Searcher()

	r := s.Search("sad")

	if len(r) != 1 {
		t.Fatalf("expected result to be 1 but was %d\n", len(r))
	}

	if r[0] != "a" {
		t.Fatal("expected doc id to be \"a\" but was", r[0])
	}
}

func TestLargeNumberOfItems(t *testing.T) {
	words := []string{"SadNanachi", "SadMario", "SadRingo", "SadNeko", "sadcat", "sadshinji", "sad_blob", "Sad_Squidward_Pepe", "SadCowblob", "SadThonk", "SadJones", "sadness", "sadgery", "sadmmLol", "sad_anger", "SadPhox", "Sad_Thor", "sadtighteyes", "SadBlobThink", "sad_gym", "sadrage"}
	negativeMatches := []string{"bad", "car", "pepe", "chill", "blobsweat"}

	b := NewBuilder()
	for _, w := range append(negativeMatches, words...) {
		b.AddDoc(DocID(w), searchable{w})
	}

	s := b.Searcher()

	res := s.Search("sad")

	if len(res) != len(words) {
		t.Fatalf("expected \"%d\" got \"%d\": %v", len(words), len(res), res)
	}
}
