package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {

	method := flag.String("m", "GET", "Request method. Can only be POST/PUT/GET/DELETE")
	n := flag.Int("n", 10, "Number of requests to be generated")
	endpoint := flag.String("e", "http://localhost:8080/function/test-db", "Endpoint to send the requests")
	contentType := flag.String("c", "text/plain", "content/type for requests. Ignored if method is GET")
	data := flag.String("d", "", "Request's data. Ignored if method is GET")
	verbose := flag.Bool("v", false, "If set, response body data will be printed")
	random := flag.Bool("r", false, "If set, data is ignored and random data is generated for each request")

	flag.Parse()

	if *method != "POST" && *method != "GET" && *method != "PUT" && *method != "DELETE" {
		fmt.Println("Usage: load_generator.go")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var wg sync.WaitGroup

	for i := 0; i < *n; i++ {
		wg.Add(1)
		go func() {
			rand.Seed(time.Now().UnixNano())

			defer wg.Done()
			var resp *http.Response
			var req *http.Request
			var err error

			if *random {
				*data = randSeq(10)
			}

			if *method == "GET" {
				resp, err = http.Get(*endpoint)
				if err != nil {
					log.Fatalln("Something went wrong")
				}
			} else if *method == "POST" {
				resp, err = http.Post(*endpoint, *contentType, bytes.NewBuffer([]byte(*data)))
				if err != nil {
					log.Fatalln("Something went wrong")
				}
			} else if *method == "PUT" {
				client := &http.Client{}

				req, err = http.NewRequest(http.MethodPut, *endpoint, bytes.NewBuffer([]byte(*data)))
				if err != nil {
					log.Fatalln("Something went wrong")
				}
				req.Header.Set("Content-Type", *contentType)

				resp, err = client.Do(req)
				if err != nil {
					log.Fatalln("Something went wrong")
				}
			} else {
				client := &http.Client{}

				req, err = http.NewRequest(http.MethodDelete, *endpoint, bytes.NewBuffer([]byte(*data)))
				if err != nil {
					log.Fatalln("Something went wrong")
				}
				req.Header.Set("Content-Type", *contentType)

				resp, err = client.Do(req)
				if err != nil {
					log.Fatalln("Something went wrong")
				}
			}

			log.Printf("Request sent with response code %d\n", resp.StatusCode)
			if *verbose {
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				bodyString := string(bodyBytes)
				log.Println(bodyString)
			}
		}()
	}

	wg.Wait()
}
