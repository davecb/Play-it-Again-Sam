scp:
	-rm raw.csv
	scp -P 5678 dcollierbrown@localhost:go/src/loadTesting/cmd/runLoadTest/vt-ceph-001/raw.csv raw.csv || \
	scp dcollierbrown@vt-ceph-001:go/src/loadTesting/cmd/runLoadTest/vt-ceph-001/raw.csv raw.csv
	#scp -P 5678 dcollierbrown@localhost:Migration/kobobooks/real_time/raw.csv raw.csv
	perf2seconds raw.csv >time.csv
