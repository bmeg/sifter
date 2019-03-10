

#hack to get around submodule weirdness in automated docker builds
hub-build:
	go get ./
	go install ./
