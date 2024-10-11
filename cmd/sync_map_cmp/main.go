package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

func main() {
	s := settings{
		creators: workerSettings{
			workers: 1000,
			jobs:    10_000_000,
		},
		readers: workerSettings{
			workers: 1000,
			jobs:    100_000_000,
		},
		updaters: workerSettings{
			workers: 1000,
			jobs:    10_000_000,
		},
		deleters: workerSettings{
			workers: 1000,
			jobs:    10_000_000,
		},
	}
	checkMapSeq(s)
	/*
		create - main.wp:               3.140909082s
		read - main.wp:         		26.392834298s
		update - main.wp:               3.052722483s
		delete - main.wp:               2.624378569s
	*/

	checkMapParallel(s)
	/*
		update - main.wp2:              20.44424436s
		read - main.wp2:                28.365686037s
		delete - main.wp2:              28.782918547s
		create - main.wp2:              28.958680824s
	*/

	checkSyncMapParallel(s)
	/*
		read - main.wp2:                5.657860188s
		delete - main.wp2:              5.823668901s
		update - main.wp2:              6.290656854s
		create - main.wp2:              7.112568442s
	*/
}

type settings struct {
	creators workerSettings
	readers  workerSettings
	updaters workerSettings
	deleters workerSettings
}

type workerSettings struct {
	workers int
	jobs    int
}

func checkMapSeq(s settings) {
	m := make(map[string]int)
	mu := new(sync.RWMutex)
	wg := new(sync.WaitGroup)

	wp("create", s.creators, wg, func() {
		k := genString(3)
		mu.Lock()
		m[k]++
		mu.Unlock()
	})
	wp("read", s.readers, wg, func() {
		k := genString(3)
		mu.RLock()
		v, ok := m[k]
		mu.RUnlock()
		_, _ = v, ok
	})
	wp("update", s.updaters, wg, func() {
		k := genString(3)

		mu.RLock()
		_, ok := m[k]
		mu.RUnlock()

		if ok {
			mu.Lock()
			m[k] = rand.Intn(100)
			mu.Unlock()
		}
	})
	wp("delete", s.deleters, wg, func() {
		k := genString(3)
		mu.Lock()
		delete(m, k)
		mu.Unlock()
	})
}

func checkMapParallel(s settings) {
	m := make(map[string]int)
	mu := new(sync.RWMutex)
	wg := new(sync.WaitGroup)
	wg.Add(4)

	go wp2("create", s.creators, wg, func() {
		k := genString(3)
		mu.Lock()
		m[k]++
		mu.Unlock()
	})
	go wp2("read", s.readers, wg, func() {
		k := genString(3)
		mu.RLock()
		v, ok := m[k]
		mu.RUnlock()
		_, _ = v, ok
	})
	go wp2("update", s.updaters, wg, func() {
		k := genString(3)

		mu.RLock()
		_, ok := m[k]
		mu.RUnlock()

		if ok {
			mu.Lock()
			m[k] = rand.Intn(100)
			mu.Unlock()
		}
	})
	go wp2("delete", s.deleters, wg, func() {
		k := genString(3)
		mu.Lock()
		delete(m, k)
		mu.Unlock()
	})

	wg.Wait()
}

func checkSyncMapParallel(s settings) {
	m := sync.Map{}
	wg := new(sync.WaitGroup)
	wg.Add(4)

	go wp2("create", s.creators, wg, func() {
		k := genString(3)
		m.Store(k, rand.Intn(100))
	})
	go wp2("read", s.creators, wg, func() {
		k := genString(3)
		_, _ = m.Load(k)
	})
	go wp2("update", s.creators, wg, func() {
		k := genString(3)
		_, ok := m.Load(k)
		if ok {
			m.Store(k, rand.Intn(100))
		}
	})
	go wp2("delete", s.creators, wg, func() {
		k := genString(3)
		m.Delete(k)
	})

	wg.Wait()
}

func wp(name string, s workerSettings, wg *sync.WaitGroup, fn func()) {
	start := time.Now()
	jobs := make(chan struct{}, s.workers)
	wg.Add(s.workers)
	for wIdx := 1; wIdx <= s.workers; wIdx++ {
		go func() {
			defer wg.Done()
			for range jobs {
				fn()
			}
		}()
	}

	for j := 0; j < s.jobs; j++ {
		jobs <- struct{}{}
	}
	close(jobs)
	wg.Wait()
	bench(name, 1, start)
}

func wp2(name string, s workerSettings, wg *sync.WaitGroup, fn func()) {
	start := time.Now()
	jobs := make(chan struct{}, s.workers)
	for wIdx := 1; wIdx <= s.workers; wIdx++ {
		go func() {
			for range jobs {
				fn()
			}
		}()
	}

	for j := 0; j < s.jobs; j++ {
		jobs <- struct{}{}
	}
	close(jobs)
	wg.Done()
	bench(name, 1, start)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func genString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func bench(msg string, skip int, start time.Time) {
	pc, _, _, _ := runtime.Caller(skip)
	fmt.Printf("%s - %s:\t\t%s\n", msg, runtime.FuncForPC(pc).Name(), time.Since(start))
}
