.PHONY: build testacc fmtcheck

build: fmt
	go install

testacc:
	TF_ACC=1 go test ./...

fmt:
	go run mvdan.cc/gofumpt -w ./

check:
	go run honnef.co/go/tools/cmd/staticcheck ./...
	go run golang.org/x/vuln/cmd/govulncheck -v ./...
