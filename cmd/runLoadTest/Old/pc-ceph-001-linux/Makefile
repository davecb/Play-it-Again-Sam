#
# try migrate  (read) tests against ceph->mogile. Assumes we've 
# 	a tunnel to port 8080
#
TYPE=--rest
CONTENT=http://localhost:8080/content/v1/images.s3.kobo.com
IMAGES=http://localhost:8080/images/v1/images.s3.kobo.com

TPS=100
FROM=0
#FOR=1
DURATION=10
PROGRESSION=10

all: progression

step: install
	runLoadTest ${OPTS} --tps ${TPS} -from ${FROM} -for ${FOR} \
		./migrate.perf_csv \
		${CONTENT} >raw.csv
	perf2seconds raw.csv >step.csv

progression: install
	runLoadTest ${OPTS} --tps ${TPS} --progress 10 --from ${FROM} \
		--duration ${DURATION} ./migrate.perf_csv \
		${CONTENT} >raw.csv
	perf2seconds raw.csv >progression.csv

install:
	go install ../../../pkg/loadTesting
	go install ../../runLoadTest

