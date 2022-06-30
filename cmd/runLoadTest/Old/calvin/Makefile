HOST=calvin
TYPE=--rest

# The classic progressive test, of various sizes
progression: install
	runLoadTest ${TYPE}  --tps 400 --progress 10 --duration 10 \
		./load.csv http://${HOST} > raw.csv
	perf2seconds raw.csv >new_z_2.csv

probe: install
	runLoadTest -v ${TYPE}  --tps 1 --for 1 \
		./load.csv http://${HOST}
put: install
	runLoadTest -v -d ${TYPE} -rw 1024 --tps 1 --for 1 \
		./put.csv http://${HOST} 2>&1 | more

tail: install
	@echo "now cat ../load.csv >>tail.csv"
	runLoadTest -tail  ${TYPE}  --tps 100 \
		tail.csv http://${HOST}

install:
	go install github.com/davecb/Play-it-Again-Sam/pkg/loadTesting
	go install github.com/davecb/Play-it-Again-Sam/cmd/runLoadTest
