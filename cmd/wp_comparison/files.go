package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
	"unicode"
)

func generateFiles(base string) {
	generateUrls(path.Join(base, "urls_10.txt"), 10, true)
	generateUrls(path.Join(base, "urls_100.txt"), 100, true)
	generateUrls(path.Join(base, "urls_1_000.txt"), 1_000, true)
	generateUrls(path.Join(base, "urls_10_000.txt"), 10_000, true)
	generateUrls(path.Join(base, "urls_100_000.txt"), 100_000, true)
	generateUrls(path.Join(base, "urls_1_000_000.txt"), 1_000_000, true)
	generateUrls(path.Join(base, "urls_10_000_000.txt"), 10_000_000, true)
}

func generateUrls(file string, count int, showResults bool) error {
	urls := fakeUrls(count, "http://localhost:8087/status/", true)

	now := time.Now()
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for _, u := range urls {
		_, err = writer.WriteString(u + "\n")
		if err != nil {
			return fmt.Errorf("write to file: %w", err)
		}
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("flush writer: %w", err)
	}
	if showResults {
		log.Printf("file generated in %s\n", time.Since(now))
	}

	return nil
}

func readUrls(file string, showResults bool) []string {
	start := time.Now()

	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("open file to read: %v\n", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var digits []rune
	for _, char := range file {
		if unicode.IsDigit(char) {
			digits = append(digits, char)
		}
	}

	var lines []string
	if len(digits) != 0 {
		count, _ := strconv.Atoi(string(digits))
		lines = make([]string, count)
	}

	for idx := 0; idx < len(lines); idx++ {
		scanner.Scan()
		lines[idx] = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scanner err: %v\n", err)
	}

	if showResults {
		log.Printf("file read in %s\n", time.Since(start))
	}

	return lines
}

func fakeUrls(count int, baseUrl string, showResults bool) []string {
	start := time.Now()
	urls := make([]string, count)
	for idx := 0; idx < len(urls); idx++ {
		respCode := rand.Intn(500) + 200
		path, _ := url.JoinPath(baseUrl, strconv.Itoa(respCode))
		urls[idx] = path
	}
	if showResults {
		log.Printf("fake url generated in %s\n", time.Since(start))
	}

	return urls
}
