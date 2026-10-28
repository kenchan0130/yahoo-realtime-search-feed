package repositories

import (
	"regexp"
	"testing"
)

var enclosureRegexp = regexp.MustCompile(`\tSTART\t(.*?)\tEND\t`)

func normalizeEnclosureByRegexp(s string) string {
	return enclosureRegexp.ReplaceAllStringFunc(s, func(m string) string {
		if matches := enclosureRegexp.FindStringSubmatch(m); len(matches) == 2 && matches[1] != "" {
			return matches[1]
		}
		return m
	})
}

func FuzzNormalizeEnclosure(f *testing.F) {
	for _, s := range []string{
		"",
		"plain text",
		"a\tSTART\tword\tEND\tb",
		"\tSTART\tx\tEND\t\tSTART\ty\tEND\t",
		"\tSTART\t\tEND\tz",
		"\tSTART\t\tSTART\tw\tEND\t",
		"\tSTART\tline\nbreak\tEND\t",
		"\tSTART\tno end",
		"\tSTART\ta\n\tSTART\tb\tEND\t",
		"日本語\tSTART\t検索\tEND\tです",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		want := normalizeEnclosureByRegexp(s)
		if got := *normalizeEnclosure(&s); got != want {
			t.Errorf("normalizeEnclosure(%q) = %q, want %q", s, got, want)
		}
	})
}
