# build creates a local binary, for you to use
# install puts it in the default golang bin directory
# systemd tells you how to run it continuously
systemd:
	@echo "As root, run"
	@echo "cp dummy.service /usr/lib/systemd/system/dummy.service"
	@echo "systemctl start dummy"
	@echo "systemctl enable dummy"
	@echo "systemctl status dummy"

install: build
	go install github.com/davecb/Play-it-Again-Sam/cmd/dummy

build:
	go build


