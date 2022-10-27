.PHONY: build testacc fmtcheck

build: fmtcheck
	go install

testacc:
	TF_ACC=1 go test ./...

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

