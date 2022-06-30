IMAGES=http://localhost:5280//images.s3.kobo.com
DOWNLOAD=http://localhost:5280/download.s3.kobo.com
#IMAGES=http://localhost:5280/images.s3.kobo.com
#DOWNLOAD=http://10.92.10.201:5280/download.s3.kobo.com
#IMAGES=http://10.92.10.201:5280/images.s3.kobo.com

probe: install
	runLoadTest -d -v --rest  --tps 1 --for 1 \
		./probe.csv ${DOWNLOAD} 

download: install
	 runLoadTest --rest --tps 1 \
		 ./download_samples.csv ${DOWNLOAD}
images: install
	 runLoadTest --rest --tps 1 --for 1 \
		 ./image_samples.csv ${IMAGES}

progress: install
	runLoadTest --rest --tps 100 --progress 10 \
		./image_samples.csv ${DOWNLOAD} |\
		tee raw.csv

fail_short: install # will run out of inputs
	runLoadTest --rest --tps 100 --progress 10 \
		./download_samples.csv ${DOWNLOAD}

install:
	go install github.com/davecb/Play-it-Again-Sam/pkg/loadTesting
	go install github.com/davecb/Play-it-Again-Sam/cmd/runLoadTest
