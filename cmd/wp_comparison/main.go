// Tasks:
// 1. Make a function that runs goroutine for each http request
// 2. Make same functionality but with worker pool
// 3. Benchmark results against stubs and real http request
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	genFiles := flag.Bool("g", false, "generate files with fake urls")
	genPath := flag.String("p", "/tmp/", "path for generated files, used only with -g")
	genBaseUrl := flag.String("u", "http://localhost:8087/status/", "base url for generated url, used only with -g")
	genCounts := flag.String("c", "10,100,1000,10000", "number of urls to generate, separated by comma for each file")
	testFile := flag.String("f", "/tmp/urls_10k.txt", "file with urls to test")
	flag.Parse()

	if *genFiles {
		c := strings.Split(*genCounts, ",")
		var counts []int
		for _, s := range c {
			atoi, err := strconv.Atoi(s)
			if err != nil {
				log.Fatalf("failed to parse count: %v\n", err)
			}
			counts = append(counts, atoi)
		}

		files, err := generateFiles(*genPath, *genBaseUrl, counts)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("files generated: %v\n", files)
	}

	urls := readUrls(*testFile, true)

	fmt.Printf(`### INFO
Processors:		%d
File:			%s
Lines in file:		%d
###
`, runtime.NumCPU(), *testFile, len(urls))

	stub := func(url string) int {
		return 200
	}
	realRequest := func(url string) int {
		resp, err := http.Get(url)
		if err != nil {
			return -1
		}
		return resp.StatusCode
	}

	fmt.Println("stubs")
	goEachRequest(urls, stub, true)
	workerPool(urls, 1, stub, true)
	workerPool(urls, 2, stub, true)
	workerPool(urls, 4, stub, true)
	workerPool(urls, 8, stub, true)
	workerPool(urls, 16, stub, true)
	workerPool(urls, 32, stub, true)
	workerPool(urls, 256, stub, true)
	workerPool2(urls, 1, stub, true)
	workerPool2(urls, 2, stub, true)
	workerPool2(urls, 4, stub, true)
	workerPool2(urls, 8, stub, true)
	workerPool2(urls, 16, stub, true)
	workerPool2(urls, 32, stub, true)
	workerPool2(urls, 256, stub, true)

	fmt.Printf("\nreal requests\n")
	goEachRequest(urls, realRequest, true)
	workerPool(urls, 1, realRequest, true)
	workerPool(urls, 2, realRequest, true)
	workerPool(urls, 4, realRequest, true)
	workerPool(urls, 8, realRequest, true)
	workerPool(urls, 16, realRequest, true)
	workerPool(urls, 32, realRequest, true)
	workerPool(urls, 256, realRequest, true)
	workerPool2(urls, 1, realRequest, true)
	workerPool2(urls, 2, realRequest, true)
	workerPool2(urls, 4, realRequest, true)
	workerPool2(urls, 8, realRequest, true)
	workerPool2(urls, 16, realRequest, true)
	workerPool2(urls, 32, realRequest, true)
	workerPool2(urls, 256, realRequest, true)
}

// goEachRequest creates a goroutine for each element of urls to get response code.
func goEachRequest(urls []string, reqFn func(url string) int, showResults bool) map[int]int {
	start := time.Now()
	result := make(map[int]int)
	mu := sync.Mutex{}
	var wg sync.WaitGroup

	for _, u := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			code := reqFn(url)

			mu.Lock()
			result[code]++
			mu.Unlock()
		}(u)
	}
	wg.Wait()
	if showResults {
		bench(1, start)
	}

	return result
}

// workerPool makes requests using worker pool:
// - Result calculated inside each worker
// - Map access synced with mutex
func workerPool(urls []string, workerCount int, reqFn func(url string) int, showResults bool) map[int]int {
	start := time.Now()
	result := make(map[int]int)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	jobs := make(chan string, workerCount)

	// run workers
	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go func(jobs chan string) {
			defer wg.Done()
			for job := range jobs {
				code := reqFn(job)

				mu.Lock()
				result[code]++
				mu.Unlock()
			}
		}(jobs)
	}

	// send jobs
	for _, u := range urls {
		jobs <- u
	}
	close(jobs)

	wg.Wait()
	if showResults {
		benchWorker(1, start, workerCount)
	}

	return result
}

// workerPool2 makes requests using worker pool too, but:
// - Jobs channel is larger
// - Result calculated by 'results' channel
// - Map access sync is not required
func workerPool2(urls []string, workerCount int, reqFn func(url string) int, showResults bool) map[int]int {
	start := time.Now()
	result := make(map[int]int)
	jobsCount := len(urls)
	jobs := make(chan string, workerCount)
	results := make(chan int, jobsCount)

	// run workers
	for w := 1; w <= workerCount; w++ {
		go func(jobs chan string) {
			for job := range jobs {
				results <- reqFn(job)
			}
		}(jobs)
	}

	// send jobs
	for _, u := range urls {
		jobs <- u
	}
	close(jobs)

	// get results
	for i := 0; i < jobsCount; i++ {
		result[<-results]++
	}

	if showResults {
		benchWorker(1, start, workerCount)
	}

	return result
}

func bench(skip int, start time.Time) {
	pc, _, _, _ := runtime.Caller(skip)
	fmt.Printf("%s:\t\t%s\n", runtime.FuncForPC(pc).Name(), time.Since(start))
}

func benchWorker(skip int, start time.Time, workerCount int) {
	pc, _, _, _ := runtime.Caller(skip)
	fmt.Printf("%s [%d]:\t\t%s\n", runtime.FuncForPC(pc).Name(), workerCount, time.Since(start))
}
