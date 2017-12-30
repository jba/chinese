package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var dataFile = flag.String("i", "", "file of data")

var input *bufio.Reader

type Item struct {
	English    string
	Pinyin     string
	Characters string
}

func main() {
	flag.Parse()
	if *dataFile == "" {
		log.Fatal("need -i")
	}
	items, err := readDataFile(*dataFile)
	if err != nil {
		log.Fatal(err)
	}
	input = bufio.NewReader(os.Stdin)

	quiz(items)
}

func readDataFile(filename string) ([]*Item, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var items []*Item
	s := bufio.NewScanner(f)
	n := 0
	for s.Scan() {
		n++
		line := strings.TrimSpace(s.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Split(s.Text(), "\t")
		if len(parts) != 3 {
			return nil, fmt.Errorf("%s:%d: need 3 parts", filename, n)
		}
		items = append(items, &Item{
			English:    parts[0],
			Pinyin:     parts[1],
			Characters: parts[2],
		})
	}
	if s.Err() != nil {
		return nil, s.Err()
	}
	return items, nil
}

func quiz(items []*Item) {
	unfinished := map[*Item]bool{}
	for _, i := range items {
		unfinished[i] = true
	}

	for len(unfinished) > 0 {
		fmt.Printf("%d items to study.\n", len(unfinished))
		for item := range unfinished {
			fmt.Printf("%s  ", item.English)
			input.ReadString('\n')
			fmt.Printf("%s\t", item.Pinyin)
			if yorn("y/n? ") {
				delete(unfinished, item)
			}
		}
		fmt.Println()
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
