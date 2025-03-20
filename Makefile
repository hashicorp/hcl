fmtcheck:
	"$(CURDIR)/scripts/gofmtcheck.sh"

fmtfix:
	gofmt -w ./

vetcheck:
	go vet ./...

copyrightcheck:
	go run github.com/hashicorp/copywrite@latest headers --plan

copyrightfix:
	go run github.com/hashicorp/copywrite@latest headers

test:
	go test ./... -race

check: copyrightcheck vetcheck fmtcheck

fix: copyrightfix fmtfix
