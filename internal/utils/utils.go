package utils

import (
	"log"
	"strings"
	"unicode"
)

func CheckEnv(value string, envParam string) string {
	if value == "" {
		log.Fatalf("%s must be set", envParam)
	}
	return value
}

func SplitString(s, delimeter string) []string {

	return strings.Split(s, delimeter)

}

func GetKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func IsMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

func RemoveSpacesAndDots(s string) string {

	s = strings.ReplaceAll(s, " ", "")

	s = strings.ReplaceAll(s, ".", "")

	return s
}

func StringPrepareForComparision(s string) string {
	s = strings.ToUpper(s)
	s = strings.TrimSpace(s)
	return s
}
