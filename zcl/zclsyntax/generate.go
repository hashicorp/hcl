package zclsyntax

//go:generate go run expression_vars_gen.go
//go:generate ragel -Z scan_tokens.rl
//go:generate gofmt -w scan_tokens.go
