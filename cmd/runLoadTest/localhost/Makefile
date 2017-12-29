SERVER=http://localhost:5280

probe: install
	runLoadTest -d -v --rest  --tps 1 --for 1 \
		./samples.csv ${SERVER} 

run : install
	 runLoadTest --rest --tps 1 \
		 ./samples.csv ${SERVER}

progress: install
	runLoadTest --rest --tps 100 --progress 10 \
		./samples.csv ${SERVER} |\
		tee raw.csv
	perf2seconds raw.csv >progress.csv

install: ${HOME}/go/bin/sim 
	go install github.com/davecb/Play-it-Again-Sam/pkg/loadTesting
	go install github.com/davecb/Play-it-Again-Sam/cmd/runLoadTest

${HOME}/go/bin/sim: 
	@echo go get github.com/davecb/Simul-Atque
	@echo cd ${GOPATH}/src/github.com/davecb/Simul-Atque/sim
	@echo make install
