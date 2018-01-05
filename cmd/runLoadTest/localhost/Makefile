SERVER=http://localhost:5280
TPS=1

probe: install
	runLoadTest -d -v --rest  --tps 1 --for 1 \
		./samples.csv ${SERVER} 

run : install
	 runLoadTest --rest --tps ${TPS} --for ${TPS}0 \
		 ./samples.csv ${SERVER} > raw.csv
	perf2seconds raw.csv >loadgen.csv

progress: install
	runLoadTest --rest --tps 24 --progress 2 \
		./samples.csv ${SERVER} > raw.csv
	perf2seconds raw.csv >progress.csv

install: ${HOME}/go/bin/sim 
	go install github.com/davecb/Play-it-Again-Sam/pkg/loadTesting
	go install github.com/davecb/Play-it-Again-Sam/cmd/runLoadTest

${HOME}/go/bin/sim: 
	@echo "if you're going to use sim,"
	@echo "go get github.com/davecb/Simul-Atque"
	@echo "cd ${GOPATH}/src/github.com/davecb/Simul-Atque/sim"
	@echo "make install"
	@echo "or comment out this message"
