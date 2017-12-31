package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jba/chinese/study"
)

var (
	itemFile    = flag.String("i", "items.txt", "file of items")
	lexiconFile = flag.String("l", "lexicon.txt", "file of words")
	nItems      = flag.Int("n", 10, "number of items to study/quiz")
	quiz        = flag.Bool("q", false, "quiz mode")
)

var input *bufio.Reader

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
	lexicon, err := readLexicon(*lexiconFile)
	if err != nil {
		log.Fatal(err)
	}
	input = bufio.NewReader(os.Stdin)

	qi := study.BuildEntries(items, lexicon, *nItems)
	if *quiz {
		runQuiz(qi)
	} else {
		runFlashcards(qi)
	}
}

func runFlashcards(entries []*study.Entry) {
	unfinished := map[*study.Entry]bool{}
	for _, i := range entries {
		unfinished[i] = true
	}

	for len(unfinished) > 0 {
		fmt.Printf("%d items to study.\n", len(unfinished))
		var es []*study.Entry
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

func present(prefix string, e *study.Entry) bool {
	fmt.Printf("%s%-30s ", prefix, e.Question)
	input.ReadString('\n')
	fmt.Printf("%-30s ", e.Answer)
	return yorn("y/n? ")
}

func runQuiz(entries []*study.Entry) {
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

func readDataFile(filename string) ([]*study.Item, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return study.ParseItems(string(data))
}

func readLexicon(filename string) (map[string][]*study.Word, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return study.ParseLexicon(string(data))
}
