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
)

func generateFiles(folder, baseUrl string, counts []int) ([]string, error) {
	files := make([]string, len(counts))
	for idx, count := range counts {
		generatedFile, err := generateUrls(fileName(folder, count), baseUrl, count, true)
		if err != nil {
			return nil, err
		}
		files[idx] = generatedFile
	}

	return files, nil
}

func generateUrls(file, baseUrl string, count int, showResults bool) (string, error) {
	urls := fakeUrls(count, baseUrl, true)

	now := time.Now()
	f, err := os.Create(file)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for _, u := range urls {
		_, err = writer.WriteString(u + "\n")
		if err != nil {
			return "", fmt.Errorf("write to file: %w", err)
		}
	}

	err = writer.Flush()
	if err != nil {
		return "", fmt.Errorf("flush writer: %w", err)
	}
	if showResults {
		log.Printf("file generated in %s\n", time.Since(now))
	}

	return file, nil
}

func readUrls(file string, showResults bool) []string {
	start := time.Now()

	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("open file to read: %v\n", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
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
		p, _ := url.JoinPath(baseUrl, strconv.Itoa(respCode))
		urls[idx] = p
	}
	if showResults {
		log.Printf("fake url generated in %s\n", time.Since(start))
	}

	return urls
}

func fileName(folder string, count int) string {
	var file string
	for count > 1000 {
		file += "k"
		count /= 1000
	}
	file = "urls_" + strconv.Itoa(count) + file + ".txt"

	return path.Join(folder, file)
}
