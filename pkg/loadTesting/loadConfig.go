package loadTesting

// Load a config file in "name value" format

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
)

// S3config is an api key and a secret, for s3 and similar servers.
type S3config struct {
	AccessKey, SecretKey string
}

// S3params is the exported key and secret
var S3params S3config

// LoadConfig loads the s3/ceph parameters from a plain-text file.
// The format is specific to a particular customer.
func LoadConfig(fullPath string) error {
	log.Printf("in LoadLoadtestConfig(%s)\n", fullPath)

	in, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer in.Close() // nolint

	r := csv.NewReader(in)
	r.Comma = ' '
	r.Comment = '#'
	r.FieldsPerRecord = -1 // ignore differences

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("Fatal error reading %s: %s", fullPath, err)
		}
		//log.Printf("name=%s,value=%s\n", record[0], record[1])
		switch record[0] {
		case "S3_ACCESS_KEY":
			S3params.AccessKey = record[1]
		case "S3_SECRET_KEY":
			S3params.SecretKey = record[1]
		}
	}
	//if verbose {
	log.Printf("access key=%q,secret=%q\n", S3params.AccessKey, S3params.SecretKey)
	//}
	return nil
}
