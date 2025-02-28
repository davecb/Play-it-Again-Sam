#
# try out the logrocket.go limiter
#
SERVICETIME=0.1
curl: 
	for i in 1 2 3 4 5 6; do curl http://localhost:9990/; done
	@echo ""

#build:
#	go build
#install:
#	go install github.com/davecb/Play-it-Again-Sam/cmd/dummy/dummy
#
#run:
#	go run dummy.go -s ${SERVICETIME} 
#

#
# systemd comands
#
update: build
	sudo systemctl daemon-reload
	sudo systemctl restart dummy
	sudo systemctl status dummy | cat
	curl http://localhost:9990/greet/
	@echo ""

down:
	sudo systemctl daemon-reload
	sudo systemctl stop dummy
	sudo systemctl disable dummy
	-sudo systemctl status dummy | cat
	-curl http://localhost:9990/greet/
	@echo ""
	@echo "the above should show it is down"
	
