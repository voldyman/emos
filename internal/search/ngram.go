package search

import (
	"strings"
	"unicode/utf8"
)

type ngram uint64

const ngramSize = 3

func generateNgrams(input string) []ngram {
	str := []byte(strings.ToLower(input))
	var runeGram [3]rune

	result := make([]ngram, 0, len(str))
	runeCount := 0

	for len(str) > 0 {
		r, sz := utf8.DecodeRune(str)
		str = str[sz:]
		runeGram[0] = runeGram[1]
		runeGram[1] = runeGram[2]
		runeGram[2] = r

		runeCount++

		if runeCount < ngramSize {
			continue
		}
		ng := runesToGram(runeGram)
		result = append(result, ng)
	}
	return result
}

func ngramToBytes(n ngram) []byte {
	rs := ngramToRunes(n)
	return []byte{byte(rs[0]), byte(rs[1]), byte(rs[2])}
}

const runeMask = 1<<21 - 1

func ngramToRunes(n ngram) [ngramSize]rune {
	return [ngramSize]rune{rune((n >> 42) & runeMask), rune((n >> 21) & runeMask), rune(n & runeMask)}
}

func (n ngram) String() string {
	rs := ngramToRunes(n)
	return string(rs[:])
}

func runesToGram(b [ngramSize]rune) ngram {
	return ngram(uint64(b[0])<<42 | uint64(b[1])<<21 | uint64(b[2]))
}
