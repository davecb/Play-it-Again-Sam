SERVER=http://localhost:5280/
TPS=10
 
no-op: install
	@echo "runLoadTest installed in go/bin"

probe: install
	runLoadTest -d -v --rest  --tps 1 --for 1 \
		./samples.csv ${SERVER} 

run : install
	 runLoadTest --rest --tps ${TPS} --for ${TPS}0 \
		 ./samples.csv ${SERVER} > raw.csv
	perf2seconds raw.csv >loadgen.csv

progress: install
	runLoadTest --rest --tps 50 --progress 10 \
		./samples.csv ${SERVER} > raw.csv
	perf2seconds raw.csv >progress.csv

install:
	go install github.com/davecb/Play-it-Again-Sam/pkg/loadTesting
	go install github.com/davecb/Play-it-Again-Sam/cmd/runLoadTest
build:
	go install github.com/davecb/Play-it-Again-Sam/cmd/runLoadTest

# set up the entire requrements, starting with the go compiler
setup: go libs install
	@echo "add to your .profile: "
	@echo 'export GOPATH=${HOME}/go'
	@echo 'PATH=${HOME}/bin:${HOME}/.local/bin:${GOPATH}/bin:${PATH}'

go: /usr/local/go/bin/go # the go compiler, 1.8 for linux in this case
	export GOPATH=${HOME}/go
	cd /usr/local; \
	sudo curl -O https://storage.googleapis.com/golang/go1.8.linux-amd64.tar.gz && \
	sudo tar -xvf go1.8.linux-amd64.tar.gz

# the libraries used
libs: ${HOME}/go/src/github.com/aws/aws-sdk-go/aws \
	${HOME}/go/src/gopkg.in/fsnotify.v1 \
	${HOME}/go/src/github.com/vharitonsky/iniflags

${HOME}/go/src/github.com/aws/aws-sdk-go/aws:
	go get github.com/aws/aws-sdk-go/aws

${HOME}/go/src/gopkg.in/fsnotify.v1:
	go get gopkg.in/fsnotify.v1

${HOME}/go/src/github.com/vharitonsky/iniflags:
	go get github.com/vharitonsky/iniflags

# Optional simulator to load-test
${HOME}/go/bin/sim: 
	@echo "if you're going to use sim,"
	@echo "go get github.com/davecb/Simul-Atque"
	@echo 'cd ${GOPATH}/src/github.com/davecb/Simul-Atque/sim'
	@echo "make install"
	@echo "or comment out this message"
