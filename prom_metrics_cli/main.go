package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"promMetrics/requests"
	"strconv"
	"time"
)

type list []string

func (e *list) Set(value string) error {
	*e = append(*e, value)
	return nil
}

func (e *list) String() string {
	str := "["
	for i, j := range *e {
		if i != len(*e)-1 {
			str += j + ", "
		}
	}
	str += "]"
	return str
}

func parsingError() {
	fmt.Println("Usage: main.go")
	flag.PrintDefaults()
	fmt.Println("\nNote: number of expression query path arguments have to be the same as query parameter arguments")
	os.Exit(1)
}

func ParseCommandLineArgs() map[string]interface{} {

	args := make(map[string]interface{})

	var endpoint string
	var path list
	var stringParams list
	var interval list
	flag.StringVar(&endpoint, "e", "http://localhost:9090/", "Prometheus endpoint")
	flag.Var(&path, "p", "Expression query path list (default 'api/v1/query')")
	flag.Var(&stringParams, "params", "Query parameters as json string list")
	flag.Var(&interval, "i", "If -i is set then an interval will be returned ending at current time."+
		"Overrides 'start' and 'end' parameters."+
		"If -i is found less times than 'path' and 'params' it is included in the beggining")

	flag.Parse()

	if len(path) < len(stringParams) {
		pathSize := len(path)
		for i := 0; i < len(stringParams)-pathSize; i++ {
			path = append(path, "api/v1/query")
		}
	}

	if len(path) != len(stringParams) || stringParams == nil {
		parsingError()
	}

	var convertedInterval []int
	for _, j := range interval {
		i, err := strconv.Atoi(j)
		if err != nil {
			parsingError()
		}
		convertedInterval = append(convertedInterval, i)
	}

	params := make([]map[string]string, len(stringParams))
	for i, p := range stringParams {
		json.Unmarshal([]byte(p), &params[i])
		if i < len(convertedInterval) {
			if convertedInterval[i] > 0 {
				params[i]["start"] = strconv.Itoa(int(time.Now().Unix()) - convertedInterval[i])
				params[i]["end"] = strconv.Itoa(int(time.Now().Unix()))
			}
		}
	}

	args["endpoint"] = endpoint
	args["path"] = path
	args["params"] = params

	return args
}

func main() {

	args := ParseCommandLineArgs()
	params := args["params"].([]map[string]string)
	endpoint := args["endpoint"].(string)
	path := args["path"].(list)
	var urls []string

	for i := 0; i < len(params); i += 1 {
		PROM_URL := requests.CreateQueryURL(endpoint, path[i], params[i])
		log.Println(PROM_URL)
		urls = append(urls, PROM_URL)
	}

	requests.Request(urls)
}
