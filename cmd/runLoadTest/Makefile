#
# the build steps and installation in go/bin
# FIXME there should be roughly one makefile 
#

build:  # This compiles a local copy
	# It can be used by persons who aren't interested or able to compile 
	# it. It can be used by them by copying it to a location in their PATH.
	go build

install: build # Install in go bin directory
	go install github.com/davecb/Play-it-Again-Sam/cmd/runLoadTest

