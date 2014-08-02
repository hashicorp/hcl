default: test

fmt: y.go json/y.go
	go fmt ./...

test: y.go json/y.go
	go test ./...

y.go: parse.y
	go tool yacc -p "hcl" parse.y

json/y.go: json/parse.y
	cd json/ && \
		go tool yacc -p "json" parse.y

clean:
	rm -f y.go
	rm -f json/y.go

.PHONY: default test
