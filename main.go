package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var (
	// Command-line arguments.
	filtersArg    = flag.String("filters", "{}", "A JSON string containg key-regex pairs")
	numGoroutines = flag.Int("numGoroutines", 4, "Number of concurrent goroutines to use for parsing the stream")
	project       = flag.String("project", "", "A comma-separated list of string representing keys to be printed out")
	outputFormat  = flag.String("outputFormat", "json", "csv or json")
)

type stringOnlyJSON map[string]string

func (s stringOnlyJSON) String() string {

	// Sort the keys on the map to ensure consistent output
	// TODO: Evaluate performance vs using fmt.Sprint to the string, and process
	// that instead.
	keys := make([]string, len(s))
	i := 0
	for k := range s {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	// Actually builds the output.
	var sb strings.Builder
	sb.WriteString("{")
	for idx, key := range keys {
		if idx != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("\"%s\": \"%s\"", key, s[key]))
	}
	sb.WriteString("}")
	return sb.String()

}

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

func getSerializer(format string) serializer {
	switch format {
	case "json":
		return jsonSerializer{}
	case "csv":
		return csvSerializer{}
	default:
		panic(fmt.Errorf("unknown serialization format: %s", format))
	}
}

func main() {
	flag.Parse()

	s := getSerializer(*outputFormat)

	// Set up printing queue and goroutine.
	printQueue := make(chan stringOnlyJSON)
	printerDone := make(chan bool)
	go printer(printQueue, printerDone, s)

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
	<-printerDone
}

func printer(toPrint <-chan stringOnlyJSON, done chan<- bool, s serializer) {
	defer func() {
		done <- true
	}()

	for line := range toPrint {
		if *project == "" {
			fmt.Println(s.serialize(line))
			continue
		}

		projected := stringOnlyJSON{}
		for _, key := range strings.Split(*project, ",") {
			projected[key] = line[key]
		}
		fmt.Println(s.serialize(projected))
	}
}

func parser(f Filterer, toParse <-chan string, parsed chan<- stringOnlyJSON,
	wg *sync.WaitGroup) {
	defer wg.Done()
	for line := range toParse {
		jsonObj := stringToJSON(line)
		if !f.ToKeep(jsonObj) {
			continue
		}
		parsed <- jsonObj
	}
}

// Utility Functions

func stringToJSON(s string) (j stringOnlyJSON) {
	if err := json.Unmarshal([]byte(s), &j); err != nil {
		panic(err)
	}
	return
}
