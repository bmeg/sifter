

SIFTER_VERSION=0.2.0

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
	@golangci-lint run

lint-depends:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1

test: .TEST

.TEST:
	go test ./test

