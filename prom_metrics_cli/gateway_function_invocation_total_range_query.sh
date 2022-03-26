#!/bin/bash

go run main.go -p 'api/v1/query_range' -i 30 -params '{"step": "1s", "query": "gateway_function_invocation_total"}' &>> log.txt
tail -1 log.txt | python plot/plot.py -f "code" -x "Time(s)" -y "Requests no."
