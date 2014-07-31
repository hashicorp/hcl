default: test

test: y.go
	go test

y.go: parse.y
	go tool yacc -p "hcl" parse.y

.PHONY: default test
