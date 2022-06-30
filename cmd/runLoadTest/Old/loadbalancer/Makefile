ceph: install
	runLoadTest --rest --tps 1 -v -d \
		--host-header download.kobobooks.com \
		--cache \
		--headers "albert:alligator pogo:possum" \
		probe.csv \
		http://storage-cs-download.kobobooks.com 2>&1 |\
		more
mogile: install
	runLoadTest --rest --tps 1 -v -d \
		--host-header download.kobobooks.com \
		--cache --akamai-debug \
                probe.csv \
		http://origin-download.kobobooks.com 2>&1 |\
		more

akamai: install
	runLoadTest --rest --tps 1 -v -d \
		--akamai-debug \
		akamai.csv https://kbimages1-a.akamaihd.net 2>&1 |\
		more

sample: # ask about odd stuff in sample.csv
	runLoadTest --rest --tps 1 -v -d \
		--host-header download.kobobooks.com \
		sample.csv http://storage-cs-download.kobobooks.com 2>&1 |\
                more


install:
	go install ../../../pkg/loadTesting
	go install ../../runLoadTest


