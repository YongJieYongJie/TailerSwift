package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sync"
)

var (
	// Command-line arguments.
	filtersArg    = flag.String("filters", "{}", "A JSON string containg key-regex pairs")
	numGoroutines = flag.Int("numGoroutines", 4, "Number of concurrent goroutines to use for parsing the stream")
)

type stringOnlyJSON map[string]string

type Filterer interface {
	ToKeep(jsonObj stringOnlyJSON) bool
}

// filter encapsulates filtering rules represented as a map of JSON key to
// regexp criteria.
type filter struct {
	rules map[string]string
}

func (f filter) ToKeep(jsonObj stringOnlyJSON) bool {
	for key, matcher := range f.rules {
		value, ok := jsonObj[key]
		if !ok {
			return false
		}
		if matched, err := regexp.MatchString(matcher, value); !matched || err != nil {
			return false
		}
	}
	return true
}

func main() {
	flag.Parse()

	// Set up printing queue and goroutine.
	printQueue := make(chan string)
	var wgPrinter sync.WaitGroup
	wgPrinter.Add(1)
	go printer(printQueue, &wgPrinter)

	// Set up parsing queue and multiple goroutines for parsing.
	parseQueue := make(chan string)
	var wgParsers sync.WaitGroup
	filters := stringToJSON(*filtersArg)
	for i := 0; i < *numGoroutines; i++ {
		wgParsers.Add(1)
		go parser(filter{filters}, parseQueue, printQueue, &wgParsers)
	}

	// Read from STDIN and parsing and printing
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		parseQueue <- scanner.Text()
	}

	// Wait for goroutines to finish.
	close(parseQueue)
	wgParsers.Wait()

	close(printQueue)
	wgPrinter.Wait()
}

func printer(toPrint <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for line := range toPrint {
		fmt.Println(line)
	}
}

func parser(f Filterer, toParse <-chan string, parsed chan<- string,
	wg *sync.WaitGroup) {
	defer wg.Done()
	for line := range toParse {
		if !f.ToKeep(stringToJSON(line)) {
			continue
		}
		parsed <- line
	}
}

// Utility Functions

func stringToJSON(s string) (j stringOnlyJSON) {
	if err := json.Unmarshal([]byte(s), &j); err != nil {
		panic(err)
	}
	return
}
