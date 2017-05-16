package json

import (
	"github.com/agext/levenshtein"
)

var keywords = []string{"false", "true", "null"}

// keywordSuggestion tries to find a valid JSON keyword that is close to the
// given string and returns it if found. If no keyword is close enough, returns
// the empty string.
func keywordSuggestion(given string) string {
	for _, kw := range keywords {
		dist := levenshtein.Distance(given, kw, nil)
		if dist < 3 { // threshold determined experimentally
			return kw
		}
	}
	return ""
}
