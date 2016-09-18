package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var ignoreCase bool
var invertMatch bool
var ignorePatterns []*regexp.Regexp

func init() {
	flag.BoolVar(&ignoreCase, "i", false, "search is case insensitive")
	flag.BoolVar(&invertMatch, "v", false, "invert search results")
}

func main() {
	flag.Parse()

	searchRegex := flag.Arg(0)
	searchRoot := flag.Arg(1)

	if searchRoot == "" {
		searchRoot = "."
	}

	if searchRegex == "" {
		fmt.Println("no search term")
		os.Exit(1)
	}

	if ignoreCase {
		searchRegex = "(?i)" + searchRegex
	}

	r, err := regexp.Compile(searchRegex)

	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	// ignore hidden directories and files
	addIgnorePattern("^\\.")
	addIgnorePattern("/\\.")

	loadIgnorePatterns(".gitignore")
	loadIgnorePatterns(".git/info/exclude")
	loadIgnorePatterns(".fdignore")

	pathChan := generate(searchRoot)

	// start 20 search go routines...
	var wg sync.WaitGroup
	const maxSearchers = 20
	wg.Add(maxSearchers)

	for i := 0; i < maxSearchers; i++ {
		go func() {
			search(pathChan, r)
			wg.Done()
		}()
	}

	wg.Wait()
}

func generate(searchPaths ...string) <-chan string {
	out := make(chan string)
	go func() {
		for _, searchPath := range searchPaths {
			filepath.Walk(searchPath, func(path string, file os.FileInfo, err error) error {
				if !file.IsDir() && !ignorePath(path) {
					out <- path
				}

				return nil
			})

			close(out)
		}
	}()

	return out
}

func search(in <-chan string, search *regexp.Regexp) {
	for path := range in {
		file, err := os.Open(path)

		if err != nil {
			fmt.Printf("%s: %s\n", path, err.Error())
			continue
		}

		scanner := bufio.NewScanner(file)

		lineNumber := 1
		for scanner.Scan() {
			// limited to 64KB here.
			line := scanner.Text()

			match := search.MatchString(line)

			if invertMatch {
				match = !match
			}

			if match {
				fmt.Printf("%s:%d:%s\n", path, lineNumber, line)
			}

			lineNumber++
		}

		if err = scanner.Err(); err != nil {
			fmt.Printf("%s: %s\n", path, err.Error())
		}

		file.Close()
	}
}

func addIgnorePattern(pattern string) {
	r, err := regexp.Compile(pattern)

	if err != nil {
		panic(err)
	}

	ignorePatterns = append(ignorePatterns, r)
}

func ignorePath(p string) bool {
	for _, r := range ignorePatterns {
		if r.MatchString(p) {
			return true
		}
	}

	return false
}

func loadIgnorePatterns(p string) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		// p does not exist
		return
	}

	file, err := os.Open(p)

	if err != nil {
		fmt.Printf("Unable to read %s: %s", p, err.Error())
		return
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		pattern := scanner.Text()
		pattern = strings.TrimSpace(pattern)
		pattern = strings.Replace(pattern, "*", ".*", -1)

		if pattern == "" {
			continue
		}

		if string([]rune(pattern)[0]) == "#" {
			continue
		}

		addIgnorePattern(pattern)
	}

	file.Close()
}
