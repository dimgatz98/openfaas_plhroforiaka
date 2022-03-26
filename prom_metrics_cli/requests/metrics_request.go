package requests

import (
	"io"
	"log"
	"net/http"
	"sync"
)

type PromMetrics interface {
	Request(url string) (string, int)
	CreateQueryURL(endpoint string, path string, args map[string]string) string
	makeRequest(url string, wg *sync.WaitGroup)
}

func CreateQueryURL(endpoint string, path string, args map[string]string) string {

	url := endpoint + path
	count := 0
	for value, key := range args {
		if count == 0 {
			url += "?" + value + "=" + key
			count += 1
		} else {
			url += "&" + value + "=" + key
		}
	}
	return url
}

func makeRequest(url string, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}
		log.Printf("%s\n", string(bodyBytes))
	} else {
		log.Printf("Server response with status code: %d", resp.StatusCode)
	}
}

func Request(urls []string) {
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go makeRequest(url, &wg)
	}

	wg.Wait()
}
