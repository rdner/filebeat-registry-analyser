package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

const (
	defaultBufferSize = 1024
)

type meta struct {
	// filestream has the source field in the record metadata.
	Source string `json:"source"`
}

type value struct {
	Meta meta `json:"meta"`
	// log input has the source file in the value itself.
	Source string `json:"source"`
}

type record struct {
	Key   string `json:"k"`
	Value *value `json:"v"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("No filename specified, exiting.\n\n")
		fmt.Printf("Filebeat registry log analyser.\nThe tool will scan the Filebeat registry logs and report suspicious facts. You can specify multiple registry log files and the tool will concatenate them.\n\n")

		fmt.Println("Usage:\t\tregan [file ...]")
		fmt.Println("Example:\tregan log1.json log2.json log3.json")
		os.Exit(1)
	}

	files := os.Args[1:]
	log.Printf("Given %d files: %s", len(files), strings.Join(files, ", "))

	workerCount := runtime.NumCPU()
	if len(files) < workerCount {
		workerCount = len(files)
	}
	log.Printf("Starting analysis with %d workers...", workerCount)

	analyse(files, workerCount, defaultBufferSize)
}

func analyse(filenames []string, workers int, buffer int) {
	recordsCh := make(chan record, buffer)
	filenamesCh := make(chan string)
	var wg sync.WaitGroup

	// workers for reading files
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := worker(filenamesCh, recordsCh)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	occurances := make(map[string]map[string]struct{})
	var recordCount int

	// worker for consuming records and organizing them in a data structure
	go func() {
		for record := range recordsCh {
			recordCount++
			// assume filestream
			source := record.Value.Meta.Source
			// or log input
			if source == "" {
				source = record.Value.Source
			}
			// incorrect data
			if source == "" {
				log.Printf("found an incompatible record without a source file: %+v\n", record)
				continue
			}
			if _, ok := occurances[source]; !ok {
				occurances[source] = make(map[string]struct{})
			}
			occurances[source][record.Key] = struct{}{}
		}
	}()

	for _, filename := range filenames {
		filenamesCh <- filename
	}
	close(filenamesCh)

	wg.Wait()
	close(recordsCh)

	log.Printf("Found %d records in %d files", recordCount, len(filenames))
	log.Printf("Found %d unique files in the log", len(occurances))

	log.Printf("Analysing...")

	var reported int

	for filename := range occurances {
		keys := occurances[filename]
		if len(keys) == 1 {
			continue
		}
		reported++
		var printingKeys []string
		for key := range keys {
			printingKeys = append(printingKeys, key)
		}
		log.Printf("File %s has multiple keys in the registry:\n\t%s", filename, strings.Join(printingKeys, "\n\t"))
	}

	log.Printf("Analysis is complete, %d fact(s) reported", reported)
}

func worker(filenamesCh <-chan string, recordsCh chan<- record) error {
	for filename := range filenamesCh {
		log.Printf("Reading from %s...", filename)
		err := readRegistry(filename, recordsCh)
		if err != nil {
			return fmt.Errorf("failed to read records from %s: %w", filename, err)
		}
		log.Printf("Reading from %s finished.", filename)
	}
	return nil
}

func readRegistry(filename string, recordsCh chan<- record) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open a file %s: %w", filename, err)
	}

	decoder := json.NewDecoder(file)

	for {
		var next record
		err = decoder.Decode(&next)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read a record from file %s at %d: %w", filename, decoder.InputOffset(), err)
		}
		// it's not a record, it's an operation item instead (e.g. `{"op":"set","id":3923248}`)
		if next.Key == "" {
			continue
		}
		// it's a `remove` record
		if next.Value == nil {
			continue
		}

		recordsCh <- next
	}
}
