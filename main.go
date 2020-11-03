package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sync"
)

var (
	numParallel = 4
	Filters     = map[string]string{
		"request_id": ".*",
		"timestamp":  ".*",
	}
)

func main() {
	parseQueue := make(chan string)
	printQueue := make(chan string)

	var wgPrinter sync.WaitGroup
	wgPrinter.Add(1)
	go printer(printQueue, &wgPrinter)

	var wgParsers sync.WaitGroup
	for i := 0; i < numParallel; i++ {
		wgParsers.Add(1)
		go parser(parseQueue, printQueue, &wgParsers)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		parseQueue <- scanner.Text()
	}

	close(parseQueue)
	wgParsers.Wait()

	close(printQueue)
	wgPrinter.Wait()
}

func printer(lines <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for line := range lines {
		fmt.Println(line)
	}
}

func parser(in <-chan string, out chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for line := range in {
		// Parse to JSON.
		var jsonObj map[string]interface{}
		json.Unmarshal([]byte(line), &jsonObj)

		// Filter based on Filters.
		if !toKeep(jsonObj) {
			continue
		}

		out <- line
	}
}

// toKeep uses the global variable Filter to filter the jsonObj, returning true
// if for each key-regexp pair in Filter, the jsonObj satisfy the regexp for the
// corresponding key.
func toKeep(jsonObj map[string]interface{}) bool {
	for key, matcher := range Filters {
		value, ok := jsonObj[key].(string)
		if !ok {
			return false
		}
		if matched, err := regexp.MatchString(matcher, value); !matched || err != nil {
			return false
		}
	}
	return true
}
