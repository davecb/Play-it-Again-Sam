HOST=localhost:8080
TYPE=--rest

# The classic progressive test, of various sizes
progression: install
	-ulimit -Sn 524288; \
	runLoadTest ${TYPE}  --tps 100 --progress 10 --duration 10 \
		sample.csv http://${HOST} > raw.csv
	make cooked

cooked:
	perf2seconds raw.csv >cooked.csv

probe: install
	runLoadTest -v -d ${TYPE}  --tps 1 --for 1 \
		1000-row-sample.csv http://${HOST}

install:
	go install github.com/davecb/Play-it-Again-Sam/pkg/loadTesting
	go install github.com/davecb/Play-it-Again-Sam/cmd/runLoadTest

ulimit: # increase FDs
	ulimit -Sn 524288
