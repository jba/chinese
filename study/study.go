package study

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"strings"
)

type Item struct {
	English    string
	Pinyin     string
	Characters string
}

type Word struct {
	English      string
	Pinyin       string
	Characters   string
	PartOfSpeech string
}

type Entry struct {
	Question, Answer string
}

// Parse items from s.
// Each line of s consists of three tab-separated values:
// English Pinyin Characters
func ParseItems(s string) ([]*Item, error) {
	lines, err := parseTSV(s, 3)
	if err != nil {
		return nil, err
	}
	var items []*Item
	for _, parts := range lines {
		items = append(items, &Item{
			English:    parts[0],
			Pinyin:     parts[1],
			Characters: parts[2],
		})
	}
	return items, nil
}

// ParseWords parses a list of words from s. Each line of s should consist
// of four tab-separated values:
// English Pinyin PartOfSpeech Characters
func ParseWords(s string) ([]*Word, error) {
	lines, err := parseTSV(s, 4)
	if err != nil {
		return nil, err
	}
	var words []*Word
	for _, line := range lines {
		word := &Word{
			English:      line[0],
			Pinyin:       line[1],
			PartOfSpeech: line[2],
			Characters:   line[3],
		}
		words = append(words, word)
	}
	return words, nil
}

// Construct a set of n question-answer pairs from the given items and lexicon.
func BuildEntries(items []*Item, words []*Word, n int) []*Entry {
	lexicon := map[string][]*Word{}
	for _, w := range words {
		lexicon[w.PartOfSpeech] = append(lexicon[w.PartOfSpeech], w)
	}
	perm := rand.Perm(len(items))
	var result []*Entry
	if len(items) < n {
		n = len(items)
	}
	for i := 0; i < len(perm) && len(result) < n; i++ {
		item := items[perm[i]]
		e, err := entry(item, lexicon)
		if err != nil {
			log.Printf("BuildEntries: %v", err)
		} else {
			result = append(result, e)
		}
	}
	return result
}

func entry(item *Item, lexicon map[string][]*Word) (*Entry, error) {
	if !isTemplate(item.English) {
		return &Entry{item.English, item.Pinyin}, nil
	}
	q, bindings := instantiateTemplate(item.English, lexicon)
	a, err := applyBindings(item.Pinyin, bindings)
	if err != nil {
		return nil, err
	}
	return &Entry{q, a}, nil
}

func isTemplate(s string) bool {
	for _, w := range strings.Fields(s) {
		if strings.HasPrefix(w, ":") {
			return true
		}
	}
	return false
}

func instantiateTemplate(template string, lexicon map[string][]*Word) (string, map[string]string) {
	words := strings.Fields(template)
	var result []string
	bindings := map[string]string{}
	for _, w := range words {
		result = append(result, chooseWord(w, lexicon, bindings))
	}
	return strings.Join(result, " "), bindings
}

func chooseWord(w string, lexicon map[string][]*Word, bindings map[string]string) string {
	if !strings.HasPrefix(w, ":") {
		return w
	}
	v := w[1:]
	pos := v
	if last := v[len(v)-1]; last >= '0' && last <= '9' {
		pos = v[:len(v)-1]
	}
	choices := lexicon[pos]
	if len(choices) == 0 {
		bindings[v] = "???"
		return "???"
	} else {
		choice := choices[rand.Intn(len(choices))]
		bindings[v] = choice.Pinyin
		return choice.English
	}
}

func applyBindings(s string, bindings map[string]string) (string, error) {
	words := strings.Fields(s)
	var result []string
	for _, w := range words {
		if !strings.HasPrefix(w, ":") {
			result = append(result, w)
		} else if b, ok := bindings[w[1:]]; ok {
			if b == "" {
				return "", fmt.Errorf("empty string for %s", w[1:])
			}
			result = append(result, b)
		} else {
			return "", fmt.Errorf("applyBindings(%q): no binding for %q", s, w[1:])
		}
	}
	return strings.Join(result, " "), nil
}

// Parse a tab-separated value string.
// Require exactly itemsPerLine items per line.
// Collapse multiple adjacent tabs.
// Ignore blank lines or lines beginning with '#'.
func parseTSV(data string, itemsPerLine int) ([][]string, error) {
	var lines [][]string
	s := bufio.NewScanner(strings.NewReader(data))
	n := 0
	for s.Scan() {
		n++
		line := strings.TrimSpace(s.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.FieldsFunc(s.Text(), func(r rune) bool { return r == '\t' })
		if len(parts) != itemsPerLine {
			return nil, fmt.Errorf("line %d: need %d parts per line, got %d", n, itemsPerLine, len(parts))
		}
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		lines = append(lines, parts)
	}
	if s.Err() != nil {
		return nil, s.Err()
	}
	return lines, nil
}
