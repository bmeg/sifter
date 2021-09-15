

SIFTER_VERSION=0.1.5

#hack to get around submodule weirdness in automated docker builds
hub-build:
	go get ./
	go install ./

# ---------------------
# Code Style
# ---------------------
# Automatially update code formatting
tidy:
	@for f in $$(find . -path ./vendor -prune -o -name "*.go" -print | egrep -v "\.\/go\/|\.pb\.go|\.gw\.go|\.dgw\.go|underscore\.go|restapi"); do \
		gofmt -w -s $$f ;\
		goimports -w $$f ;\
	done;


# Run code style and other checks
lint:
	@golangci-lint run --timeout 2m --disable-all -E gofmt
	@golangci-lint run --timeout 2m --disable-all -E goimports
	@golangci-lint run --timeout 2m --disable-all -E misspell
	@golangci-lint run --timeout 2m --disable-all -E typecheck
	@golangci-lint run --timeout 2m --disable-all -E golint
	@golangci-lint run --timeout 2m --disable-all -E gosimple
	@golangci-lint run --timeout 2m --disable-all -E govet

lint-depends:
	@go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1

test: .TEST

.TEST:
	go test ./test

docs:
	@go run docschema/main.go | ./docschema/schema-to-markdown.py > Playbook.md
