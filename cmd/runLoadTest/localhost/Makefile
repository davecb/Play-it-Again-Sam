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

install: go libs
	export GOPATH=${HOME}/go
	go install github.com/davecb/Play-it-Again-Sam/pkg/loadTesting
	go install github.com/davecb/Play-it-Again-Sam/cmd/runLoadTest
	@echo "add to your .profile: "
	@echo "export export GOPATH=${HOME}/go"
	@echo PATH="$HOME/bin:$HOME/.local/bin:${GOPATH}/bin:$PATH"

go: /usr/bin/go # the go compiler, for ubuntu in this case
	sudo apt-get install golang

# the libraries used
libs: ${HOME}/go/src/github.com/aws/aws-sdk-go/aws \
	${HOME}/go/src/gopkg.in/fsnotify.v1

${HOME}/go/src/github.com/aws/aws-sdk-go/aws:
	go get github.com/aws/aws-sdk-go/aws

${HOME}/go/src/gopkg.in/fsnotify.v1:
	go get gopkg.in/fsnotify.v1

# Optional simulator to load-test
${HOME}/go/bin/sim: 
	@echo "if you're going to use sim,"
	@echo "go get github.com/davecb/Simul-Atque"
	@echo "cd ${GOPATH}/src/github.com/davecb/Simul-Atque/sim"
	@echo "make install"
	@echo "or comment out this message"
