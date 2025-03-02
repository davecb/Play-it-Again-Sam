MAX=30
build:
	go build
install: build
	go install github.com/davecb/Play-it-Again-Sam/cmd/hull

test: install
	go run hull.go -maxY ${MAX} sample.csv

