package scan

import (
	"regexp"
	"strings"
	"unicode"
)

// Normalize keeps only letters and digits (any script), lower-cased, so
// "Game™: Edition" and "game edition" compare equal.
func Normalize(s string) string {
	var b strings.Builder
	for _, c := range strings.ToLower(strings.ReplaceAll(s, "&", "and")) {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			b.WriteRune(c)
		}
	}
	return b.String()
}

var (
	// Bracketed release tags: "[FitGirl Repack]", "(v1.2.3)", "{GOG}".
	reBrackets = regexp.MustCompile(`\s*[\[\(\{][^\]\)\}]*[\]\)\}]`)
	// Trailing scene group or repacker tags: "-RUNE", "_CODEX", " - FitGirl Repack".
	reGroup = regexp.MustCompile(`(?i)[\s._-]+(rune|codex|empress|tenoke|plaza|skidrow|reloaded|cpy|flt|hoodlum|razor1911|dodi|fitgirl|elamigos|kaos|xatab|gog|repack|multi\d*|goldberg|onlinefix|online-fix)(\s*repacks?)?$`)
	// Version and build suffixes: "v1.0.3", "Build 12345", "Update 5".
	reVersion = regexp.MustCompile(`(?i)[\s._-]+(v\s?\d[\w.]*|build[\s._-]?\d+|update[\s._-]?\d+|\d+\.\d+(\.\d+)*)$`)
	reSpaces  = regexp.MustCompile(`\s+`)
)

// CleanTitle turns an installer or folder name into a game title:
// "Baldurs.Gate.3-RUNE" → "Baldurs Gate 3", "Hades [FitGirl Repack]" → "Hades".
func CleanTitle(s string) string {
	s = strings.TrimSpace(s)
	s = reBrackets.ReplaceAllString(s, "")
	// Scene-style names use dots or underscores instead of spaces.
	if !strings.Contains(s, " ") && (strings.Count(s, ".") >= 2 || strings.Count(s, "_") >= 1) {
		s = strings.NewReplacer(".", " ", "_", " ").Replace(s)
	}
	for i := 0; i < 3; i++ {
		before := s
		s = reGroup.ReplaceAllString(s, "")
		s = reVersion.ReplaceAllString(s, "")
		s = strings.TrimRight(s, " .-_")
		if s == before {
			break
		}
	}
	s = reSpaces.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// SortTitle drops leading articles so "The Sims 4" sorts under S.
func SortTitle(s string) string {
	l := strings.ToLower(strings.TrimSpace(s))
	for _, a := range []string{"the ", "a ", "an "} {
		if strings.HasPrefix(l, a) && len(l) > len(a) {
			return l[len(a):]
		}
	}
	return l
}
