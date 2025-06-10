
RUNARGS?=-verbosity debug
RUNCMD?=unionfees-server
RELOS?=$(shell go env GOOS)

VERSION?=$(shell git describe --tags --always --long --dirty)
TAG?=$(shell git describe --tags --abbrev=0)

LDFLAGS="-w -s -X 'main.appVersion=$(VERSION)'"

OUTDIR?=out


.PHONY: all build run test tidy no-dirty audit checks

all: tidy test build



build: $(OUTDIR)
	go build -v -ldflags $(LDFLAGS) -o $(OUTDIR)/ ./cmd/...


checks: tidy audit test no-dirty


test: 
	go test -v -ldflags $(LDFLAGS) ./...


tidy:
	@echo "tidy and fmt..."
	go mod tidy -v
	go fmt ./...


audit:
	@echo "running audit checks..."
	go mod verify
	go vet ./...
	go list -m all
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...


no-dirty:
	@echo "Checking git status..."
	@git diff --quiet || (echo "Git working directory is not clean" && exit 1)
	@git diff --cached --quiet || (echo "Git index is not clean" && exit 1)


run: 
	go run -ldflags $(LDFLAGS) ./cmd/$(RUNCMD) $(RUNARGS)


$(OUTDIR):
	mkdir $@
