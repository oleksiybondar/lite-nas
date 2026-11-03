package main

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// searchKey folds accents and common Latin spelling variants for place and tag lookup.
func searchKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer("æ", "ae", "œ", "oe", "ø", "o", "ð", "d", "þ", "th", "ł", "l", "ß", "ss").Replace(value)
	var result strings.Builder
	for _, char := range norm.NFD.String(value) {
		if !unicode.Is(unicode.Mn, char) {
			result.WriteRune(char)
		}
	}
	return norm.NFC.String(result.String())
}
