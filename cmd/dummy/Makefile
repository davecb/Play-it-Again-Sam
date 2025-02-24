SERVICETIME=0.1
curl: 
	curl http://localhost:9990/greet/
	@echo ""

build:
	go build
install:
	go install github.com/davecb/Play-it-Again-Sam/cmd/dummy/dummy

run:
	go run dummy.go -s ${SERVICETIME} 


#
# systemd comands
#
update: build
	sudo systemctl daemon-reload
	sudo systemctl restart dummy
	sudo systemctl status dummy | cat
	curl http://localhost:9990/greet/
	@echo ""

