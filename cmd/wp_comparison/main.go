package main

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// TODO сделать на воркер пуле с реальными урлами

const (
	file_10         = "/tmp/urls_10.txt"
	file_100        = "/tmp/urls_100.txt"
	file_1_000      = "/tmp/urls_1_000.txt"
	file_10_000     = "/tmp/urls_10_000.txt"
	file_100_000    = "/tmp/urls_100_000.txt"
	file_1_000_000  = "/tmp/urls_1_000_000.txt"
	file_10_000_000 = "/tmp/urls_10_000_000.txt"
)

func main() {
	fmt.Println(runtime.NumCPU())

	urls := readUrls(file_10_000, true)

	goEachRequestStub(urls, true)
	goEachRequestReal(urls, true)
	workerPoolStub(urls, 2, true)
	workerPoolStub(urls, 4, true)
	workerPoolStub(urls, 8, true)
	workerPoolStub(urls, 16, true)
	workerPoolStub(urls, 32, true)
	workerPoolReal(urls, 2, true)
	workerPoolReal(urls, 4, true)
	workerPoolReal(urls, 8, true)
	workerPoolReal(urls, 16, true)
	workerPoolReal(urls, 32, true)
	workerPoolStub(urls, 256, true)
	workerPoolReal(urls, 256, true)
}

type stats map[int]int

func goEachRequestStub(urls []string, showResults bool) stats {
	start := time.Now()
	result := make(map[int]int)
	mu := sync.Mutex{}
	var wg sync.WaitGroup

	for _, u := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			code := requestStub(u)

			mu.Lock()
			result[code]++
			mu.Unlock()
		}()
	}
	wg.Wait()
	if showResults {
		//log.Printf("stats done in %s\n", time.Since(start))
		fmt.Printf("%s\n", time.Since(start))
	}

	return result
}

func goEachRequestReal(urls []string, showResults bool) stats {
	start := time.Now()
	result := make(map[int]int)
	mu := sync.Mutex{}
	var wg sync.WaitGroup

	for _, u := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			code := doRequest(u)

			mu.Lock()
			result[code]++
			mu.Unlock()
		}()
	}
	wg.Wait()
	if showResults {
		fmt.Printf("%s\n", time.Since(start))
		//log.Printf("stats done in %s\n", time.Since(start))
	}

	return result
}

func workerPoolStub(urls []string, workerCount int, showResults bool) stats {
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
				code := requestStub(job)

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
		fmt.Printf("%s\n", time.Since(start))
		//log.Printf("stats done in %s\n", time.Since(start))
	}

	return result
}

func workerPoolReal(urls []string, workerCount int, showResults bool) stats {
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
				code := doRequest(job)

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
		fmt.Printf("%s\n", time.Since(start))
		//log.Printf("stats done in %s\n", time.Since(start))
	}

	return result
}

func requestStub(url string) int {
	return 200
}

func doRequest(url string) int {
	resp, err := http.Get(url)
	if err != nil {
		//log.Printf("http.Get error: %v\n", err)
		return -1
	}
	return resp.StatusCode
}
