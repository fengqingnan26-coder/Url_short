package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	url         = "http://localhost:8080/api/url"
	token       = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Impla3l1bGxsQGdtYWlsLmNvbSIsInVzZXJfaWQiOjEsImV4cCI6MTc0OTUxMjA0MiwiaWF0IjoxNzQ5NDI1NjQyfQ.4Ck2c8GNub1wsvPNHRVD7LhUhskFq4OatP5-5f6mKCI"
	concurrency = 300
	totalReq    = 10000
)

type ReqBody struct {
	OriginalURL string `json:"original_url"`
	CustomCode  string `json:"custom_code"`
	Duration    int    `json:"duration"`
	UserID      int    `json:"user_id"`
}

func randString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func worker(jobs <-chan int, wg *sync.WaitGroup, mu *sync.Mutex, success *int, fail *int) {
	defer wg.Done()
	client := &http.Client{}

	for job := range jobs {
		body := ReqBody{
			OriginalURL: "https://zhuanlan.zhihu.com/p/367591714",
			CustomCode:  randString(6),
			Duration:    80,
			UserID:      1,
		}
		bs, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bs))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		start := time.Now()
		resp, err := client.Do(req)
		elapsed := time.Since(start)

		if err != nil {
			mu.Lock()
			*fail++
			mu.Unlock()
			fmt.Printf("[Job %d] Error: %v\n", job, err)
			continue
		}

		resp.Body.Close()

		mu.Lock()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			*success++
		} else {
			*fail++
		}
		mu.Unlock()

		fmt.Printf("[Job %d] Status: %d, Time: %v ms\n", job, resp.StatusCode, elapsed.Milliseconds())
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	jobs := make(chan int, totalReq)
	var wg sync.WaitGroup
	var mu sync.Mutex

	successCount := 0
	failCount := 0

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(jobs, &wg, &mu, &successCount, &failCount)
	}

	for i := 0; i < totalReq; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	fmt.Printf("All requests done. Success: %d, Fail: %d\n", successCount, failCount)
}
