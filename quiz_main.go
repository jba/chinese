package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	itemFile    = flag.String("i", "items.txt", "file of items")
	lexiconFile = flag.String("l", "lexicon.txt", "file of words")
	nItems      = flag.Int("n", 10, "number of items to study/quiz")
	quiz        = flag.Bool("q", false, "quiz mode")
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

type Entry struct {
	Question, Answer string
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
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

	qi := buildEntries(items, *nItems)
	if *quiz {
		runQuiz(qi)
	} else {
		runFlashcards(qi)
	}
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
	lines, err := readTSVFile(filename, 4)
	if err != nil {
		return nil, err
	}
	lex := map[string][]*Word{}
	for _, line := range lines {
		word := &Word{
			English:      line[0],
			Pinyin:       line[1],
			PartOfSpeech: line[2],
			Characters:   line[3],
		}
		lex[word.PartOfSpeech] = append(lex[word.PartOfSpeech], word)
	}
	return lex, nil
}

func buildEntries(items []*Item, n int) []*Entry {
	perm := rand.Perm(len(items))
	var result []*Entry
	if len(items) < n {
		n = len(items)
	}
	for i := 0; i < n; i++ {
		item := items[perm[i]]
		result = append(result, entry(item))
	}
	return result
}

func runFlashcards(entries []*Entry) {
	unfinished := map[*Entry]bool{}
	for _, i := range entries {
		unfinished[i] = true
	}

	for len(unfinished) > 0 {
		fmt.Printf("%d items to study.\n", len(unfinished))
		var es []*Entry
		for e := range unfinished {
			es = append(es, e)
		}
		for _, entry := range es {
			if present("", entry) {
				delete(unfinished, entry)
			}
		}
		fmt.Println()
	}
}

func present(prefix string, e *Entry) bool {
	fmt.Printf("%s%-30s ", prefix, e.Question)
	input.ReadString('\n')
	fmt.Printf("%-30s ", e.Answer)
	return yorn("y/n? ")
}

func runQuiz(entries []*Entry) {
	fmt.Printf("A quiz with %d questions. Let's begin!\n", len(entries))
	for {
		perm := rand.Perm(len(entries))
		score := 0
		for i, pi := range perm {
			if present(fmt.Sprintf("%d: ", i+1), entries[pi]) {
				score++
			}
		}
		fmt.Printf("You got %d out of %d, which is %d%%.\n", score, len(entries), score*100/len(entries))
		if !yorn("Take it again? ") {
			break
		}
	}
}

func entry(item *Item) *Entry {
	if !isTemplate(item.English) {
		return &Entry{item.English, item.Pinyin}
	}
	q, bindings := instantiateTemplate(item.English, lexicon)
	a := applyBindings(item.Pinyin, bindings)
	return &Entry{q, a}
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
			log.Fatalf("applyBindings(%q): no binding for %q", s, w[1:])
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
