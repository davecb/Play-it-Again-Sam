scp:
	-rm raw.csv
	scp -P 4567 dcollierbrown@localhost:go/src/loadTesting/cmd/runLoadTest/s3/raw.csv raw.csv
	#scp dcollierbrown@vt-haproxy-001:go/src/loadTesting/cmd/runLoadTest/s3/raw.csv raw.csv
	perf2seconds raw.csv >time.csv
	
probe: # just for debugging
	go install ../../../pkg/loadTesting
	go install ../../runLoadTest
	runLoadTest -v --s3  --tps 1 --for 1 \
		--config ./calvin-s3test.config \
		../load.csv http://calvin
