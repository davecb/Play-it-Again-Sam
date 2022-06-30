HOST=miles:8081/content/v1/images.s3.kobo.com/

probe: install
	runLoadTest -d -v --rest  --tps 1 --for 1 \
		./suffixes.csv http://localhost:8081/

resize_test: install
	 runLoadTest --rest --tps 1 \
		 ./suffixes.csv http://localhost:8081/

install:
	go install ../../../pkg/loadTesting
	go install ../../runLoadTest
