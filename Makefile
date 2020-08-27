.PHONY: default all build clean test fmtcheck testacc sonarqube

PKG_NAME=victorops
FILES=./...

default: test

all: clean build test

build: fmtcheck
	go build

clean:
	@echo "==> Cleaning out old builds "
	go clean
	rm -rf coverage.txt .sonar .scannerwork


fmt:
	@echo "==> Fixing source code with gofmt "
	gofmt -s -w .

lint:
	@echo "==> Checking source code against linters "
	@GOGC=30 golangci-lint run $(FILES)

fmtcheck: fmt lint

test: fmtcheck
	@echo "==> Running all tests"
	go test $(FILES) -v -timeout=30s -parallel=4 -cover

testacc: clean fmtcheck
	@echo "==> Running all tests"
	go test $(FILES) -v -timeout=30s -parallel=8 -cover -coverprofile coverage.txt

sonarqube: testacc
	docker run -it -v "${PWD}:/usr/src" sonarsource/sonar-scanner-cli
