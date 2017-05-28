package zclsyntax

//go:generate go run expression_vars_gen.go
//go:generate ragel -Z scan_token.rl
//go:generate gofmt -w scan_token.go
