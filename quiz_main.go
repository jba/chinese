package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
)

var (
	itemFile    = flag.String("i", "items.txt", "file of items")
	lexiconFile = flag.String("l", "lexicon.txt", "file of words")
)

var (
	input   *bufio.Reader
	lexicon map[string][]*Word // part of speech to word
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

func main() {
	flag.Parse()
	if *itemFile == "" {
		log.Fatal("need -i")
	}
	items, err := readDataFile(*itemFile)
	if err != nil {
		log.Fatal(err)
	}
	lexicon, err = readLexicon(*lexiconFile)
	if err != nil {
		log.Fatal(err)
	}
	input = bufio.NewReader(os.Stdin)

	quiz(items)
}

func readDataFile(filename string) ([]*Item, error) {
	lines, err := readTSVFile(filename, 3)
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

func readLexicon(filename string) (map[string][]*Word, error) {
	lines, err := readTSVFile(filename, 3)
	if err != nil {
		return nil, err
	}
	lex := map[string][]*Word{}
	for _, line := range lines {
		word := &Word{
			English:      line[0],
			Pinyin:       line[1],
			PartOfSpeech: line[2],
		}
		lex[word.PartOfSpeech] = append(lex[word.PartOfSpeech], word)
	}
	return lex, nil
}

func quiz(items []*Item) {
	unfinished := map[*Item]bool{}
	for _, i := range items {
		unfinished[i] = true
	}

	for len(unfinished) > 0 {
		fmt.Printf("%d items to study.\n", len(unfinished))
		for item := range unfinished {
			question, answer := qa(item)
			fmt.Printf("%-30s ", question)
			input.ReadString('\n')
			fmt.Printf("%-30s ", answer)
			if yorn("y/n? ") {
				delete(unfinished, item)
			}
		}
		fmt.Println()
	}
}

func qa(item *Item) (q, a string) {
	if !isTemplate(item.English) {
		return item.English, item.Pinyin
	}
	q, bindings := instantiateTemplate(item.English, lexicon)
	a = applyBindings(item.Pinyin, bindings)
	return q, a
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
	i := strings.IndexRune(v, '.')
	var pos string
	if i == -1 {
		pos = v
	} else {
		pos = v[:i]
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

func applyBindings(s string, bindings map[string]string) string {
	words := strings.Fields(s)
	var result []string
	for _, w := range words {
		if !strings.HasPrefix(w, ":") {
			result = append(result, w)
		} else if b, ok := bindings[w[1:]]; ok {
			if b == "" {
				log.Fatalf("empty string for %s", w[1:])
			}
			result = append(result, b)
		} else {
			log.Fatalf("no binding for %q", w[1:])
		}
	}
	return strings.Join(result, " ")
}

func yorn(prompt string) bool {
	fmt.Print(prompt)
	for i := 0; i < 10; i++ {
		s, err := input.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			continue
		}
		if len(s) == 0 {
			continue
		}
		switch s[0] {
		case 'y', 'Y':
			return true
		case 'n', 'N':
			return false
		}
	}
	log.Fatal("too many tries")
	return false
}

// Read a tab-separated value file.
// Require exactly itemsPerLine items per line.
// Collapse multiple adjacent tabs.
// Ignore blank lines or lines beginning with '#'.
func readTSVFile(filename string, itemsPerLine int) ([][]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines [][]string
	s := bufio.NewScanner(f)
	n := 0
	for s.Scan() {
		n++
		line := strings.TrimSpace(s.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.FieldsFunc(s.Text(), func(r rune) bool { return r == '\t' })
		if len(parts) != itemsPerLine {
			return nil, fmt.Errorf("%s:%d: need %d parts per line, got %d", filename, n, itemsPerLine, len(parts))
		}
		lines = append(lines, parts)
	}
	if s.Err() != nil {
		return nil, s.Err()
	}
	return lines, nil
}
