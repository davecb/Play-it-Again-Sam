scp:
	-rm raw.csv
	scp -P 4567 localhost:/home/kobo/go/src/newLoadTesting/cmd/runLoadTest/pc-ceph-001/raw.csv raw.csv || \
	scp pc-ceph-001:/home/kobo/go/src/newLoadTesting/cmd/runLoadTest/pc-ceph-001/raw.csv raw.csv
	perf2seconds raw.csv >time.csv

#	scp_from_mac /Users/dcollierbrown/go/src/newLoadTesting/cmd/runLoadTest/pc-ceph-001/progression.csv .
#	scp_from_mac /Users/dcollierbrown/go/src/newLoadTesting/cmd/runLoadTest/pc-ceph-001/step.csv .
#	scp_from_mac /Users/dcollierbrown/go/src/newLoadTesting/cmd/runLoadTest/pc-ceph-001/raw.csv .
#
#probe: install
#	runLoadTest --rest -v --tps 10 --for 1 ../migrate.csv http://localhost:8080/content/v1/images.s3.kobo.com
#
#install:
#	go install ../../../pkg/loadTesting
#	go install ../../runLoadTest
#
#tunnel:
#	ssh -t -2 -A -Y -L 4567:10.121.100.1:22 dcollierbrown@mac
