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

// ukrainianLatinKey matches Ukrainian spellings against Latin gallery metadata.
func ukrainianLatinKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer(
		"а", "a", "б", "b", "в", "v", "г", "h", "ґ", "g", "д", "d", "е", "e", "є", "ye",
		"ж", "zh", "з", "z", "и", "y", "і", "i", "ї", "yi", "й", "i", "к", "k", "л", "l",
		"м", "m", "н", "n", "о", "o", "п", "p", "р", "r", "с", "s", "т", "t", "у", "u",
		"ф", "f", "х", "kh", "ц", "ts", "ч", "ch", "ш", "sh", "щ", "shch", "ь", "",
		"ю", "yu", "я", "ya", "’", "", "'", "",
	).Replace(value)
	return searchKey(value)
}
