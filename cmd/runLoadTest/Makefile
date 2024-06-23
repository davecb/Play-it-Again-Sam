HOST=localhost:5280
TYPE=--rest

# The classic progressive test, of various sizes
progression: install
	runLoadTest ${TYPE}  --tps 100 --progress 10 --duration 10 \
		./load.csv http://${HOST} > raw.csv
	perf2seconds raw.csv >cooked.csv

probe: install
	runLoadTest -v ${TYPE}  --tps 1 --for 1 \
		./load.csv http://${HOST}
install:
	go install github.com/davecb/Play-it-Again-Sam/pkg/loadTesting
	go install github.com/davecb/Play-it-Again-Sam/cmd/runLoadTest
