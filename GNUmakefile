.PHONY: build
build: fmt
	go install

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./...

.PHONY: fmt
fmt:
	go run mvdan.cc/gofumpt -w ./

.PHONY: check
check:
	go run honnef.co/go/tools/cmd/staticcheck ./...
	go run golang.org/x/vuln/cmd/govulncheck -v ./...
