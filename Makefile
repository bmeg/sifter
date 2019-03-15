

#hack to get around submodule weirdness in automated docker builds
hub-build:
	go get ./
	go install ./

# ---------------------
# Code Style
# ---------------------
# Automatially update code formatting
tidy:
	@for f in $$(find . -path ./vendor -prune -o -name "*.go" -print | egrep -v "\.pb\.go|\.gw\.go|\.dgw\.go|underscore\.go"); do \
		gofmt -w -s $$f ;\
		goimports -w $$f ;\
	done;


# Run code style and other checks
lint:
	@go get github.com/alecthomas/gometalinter
	@gometalinter --install > /dev/null
	@gometalinter --disable-all --enable=vet --enable=golint --enable=gofmt --enable=misspell \
		--vendor \
		-e '.*bundle.go' -e ".*pb.go" -e ".*pb.gw.go" -e ".*pb.dgw.go" -e "underscore.go" \
		./...
