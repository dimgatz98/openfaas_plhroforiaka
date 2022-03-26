## Multithreaded http load generator 
This project is written in go 1.17

```bash
go run load_generator.go --help
```
example usage for sending 20 POST requests with content/type = "text/plain" and random data in each request:
```bash
go run load_generator.go -e "http://localhost:8080/function/test-db" -m POST -c "text/plain" -v -r -n 20
```