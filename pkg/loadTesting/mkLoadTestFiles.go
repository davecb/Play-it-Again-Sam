package loadTesting

// Create files from a load script in "perf" format. This supports a program
// that GETs, not PUTs, POSTs or DELETEs. PUTs are easy, as are DELETEs,
// but POSTs are ambiguous.
// input looks like "01-Mar-2017 16:00:00 0 0 0 0 path 200 GET"

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// MkLoadTestFiles interprets the time period and decodes what to create .
func MkLoadTestFiles(f *os.File, filename, baseURL string, startFrom, runFor int, conf Config) {

	// get settings from conf parameter
	setConf(conf)
	if debug {
		log.Printf("in MkLoadTestFiles(f *os.File, filename=%s, baseURL=%s, startFrom=%d, runFor=%d)",
			filename, baseURL, startFrom, runFor)
	}

	doPrepWork(baseURL)
	log.Printf("starting...\n")
	defer os.Remove(junkDataFile) // nolint

	r := csv.NewReader(f)
	r.Comma = ' '
	r.Comment = '#'
	r.FieldsPerRecord = -1 // ignore differences

	//skip forward if startFrom is non-zero
	for i := 0; i < startFrom; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Fatal error skipping forward in %s: %s\n", filename, err)
		}
		log.Printf("skipped %s\n", record)
	}

	for i := 0; i < runFor; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Fatal error mid-way in %s: %s\n", filename, err)
		}
		log.Printf("read %s\n", record)

		// record-type logic:
		if record[pathField] == "/" {
			// not a valid file
			log.Print("ignore a request to create the root dir, /\n")
			continue
		}
		bytes := record[bytesField]
		path := record[pathField]
		returnCode := record[returnCodeField]
		operatorValue := record[operatorField]
		switch operatorValue {
		case "PUT", "POST":
			// Don't do files that will be created in the test
			log.Printf("ignore %s operation on %s\n", operatorValue, path)
			continue
		case "DELETE", "DELE":
			// Right now, create a 1-byte file to cause directory traversals.
			mkFile(baseURL, filename, path, "1")
			continue
		case "GET", "":
			// Treat as get if there is no operator supplied
			rc, err := strconv.Atoi(returnCode)
			if err != nil {
				rc = 0
			}
			shortDescr, create := codeDescr(rc)
			if create {
				log.Printf("%s, create file %s of %s bytes\n", shortDescr, path, bytes)
				mkFile(baseURL, filename, path, bytes)
			} else {
				log.Printf("%s, ignore %s\n", shortDescr, path)
			}
		}
	}
}

// mkfile creates a file of specified size or says why not.
func mkFile(baseURL, sourceFile, fullPath, size string) {
	var err error

	if debug {
		log.Printf("in mkFile(baseURL=%s, sourceFile=%s, fullPath=%s, size=%s", baseURL, sourceFile, fullPath, size)
	}
	fileSize, err := strconv.ParseInt(size, 10, 64) // FIXME hoist
	if err != nil {
		log.Fatalf("can't get size from %q", size)
	}
	switch protocol {
	case FilesystemProtocol: // prepend current directory to path
		err = TimedCreateFilesystemFile("./"+strings.TrimPrefix(fullPath, "/"), fileSize)
	case S3Protocol:
		//err =  MinioS3Put(baseURL, fullPath, size)
		err = AmazonS3Put(baseURL, fullPath, fileSize)
	case HTTPProtocol:
		RestPut("http://"+baseURL, fullPath, fileSize)
	case CephProtocol: // Pre-alpha stage
		err = createCephFile(baseURL+fullPath, fileSize)
	}
	if err != nil {
		log.Fatalf(`Fatal error mid-way in %s: "%s" while creating %s of size %s\n`,
			sourceFile, err, fullPath, size)
	}
}

// createCephFile implements a file creation on ceph, natively: refactor
func createCephFile(fullPath string, size int64) error {
	return fmt.Errorf("createCephFile(%s, %d) not implemented", fullPath, size)
}
