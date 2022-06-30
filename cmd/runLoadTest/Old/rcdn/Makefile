scp:
	-rm raw.csv
	#scp -P 4567 localhost:/home/kobo/go/src/newLoadTesting/cmd/runLoadTest/s3/raw.csv raw.csv
	scp kobo@pc-ceph-001:/home/kobo/go/src/newLoadTesting/cmd/runLoadTest/pc-ceph-001/raw.csv raw.csv
	perf2seconds raw.csv >time.csv

