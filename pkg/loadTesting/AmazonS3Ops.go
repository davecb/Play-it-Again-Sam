package loadTesting

// AmazonS3Ops implements s3 git, put and delete using the Amazon client library.
// Initially the Amazon library was too buggy, but Marcus Watt of the ceph
// team debugged it. I expect most people will use the Amazon library.

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	// FIXME aws "github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

var svc *s3.S3
var awsLogLevel = aws.LogOff
var bucket = "images.s3.kobo.com" // FIXME

// AmazonS3Get does a GET from an s3Protocol target and times it,
func AmazonS3Get(baseURL, path string) {
	log.Printf("in AmazonS3Get(%s, %s)\n", baseURL, path)

	file, err := ioutil.TempFile("/tmp", "loadTesting")
	if err != nil {
		log.Fatal("Unable to create a temp file")
	}
	defer os.Remove(file.Name()) // nolint

	downloader := s3manager.NewDownloaderWithClient(svc)
	initial := time.Now() //              				***** Response time starts
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(path),
		})
	responseTime := time.Since(initial) // 				***** Response time ends
	if err == nil {
		fmt.Printf("%s %f 0 0 %d %s 200 GET\n",
			initial.Format("2006-01-02 15:04:05.000"),
			responseTime.Seconds(), numBytes, path)
	} else {
		fmt.Printf("%s %f 0 0 %d %s 4XX GET\n",
			initial.Format("2006-01-02 15:04:05.000"),
			responseTime.Seconds(), numBytes, path)
		// Extract and report the failure, iff possible
		reqerr, ok := err.(awserr.RequestFailure)
		if ok {
			log.Printf("s3 reported an error fetching %s, %v\n",
				path, reqerr)
		}
	}
	alive <- true
}

// AmazonS3Put puts a file and times it
// error return is used only by mkLoadTestFiles
func AmazonS3Put(baseURL, path string, size int64) error {
	log.Printf("in AmazonS3Put(%s, %s, %d)\n", baseURL, path, size)

	file, err := os.Open(junkDataFile)
	if err != nil {
		return fmt.Errorf("Unable to open junk-data file %s, %v", junkDataFile, err)
	}
	defer file.Close() // nolint
	lr := io.LimitReader(file, size)

	if svc == nil {
		return fmt.Errorf("missing service %v", svc)
	}
	uploader := s3manager.NewUploaderWithClient(svc)
	initial := time.Now() //              				***** Response time starts
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		Body:   lr,
	})
	responseTime := time.Since(initial) // 				***** Response time ends
	// FIXME swap this around
	if err == nil {
		fmt.Printf("%s %f 0 0 %d %s 201 PUT\n",
			initial.Format("2006-01-02 15:04:05.000"),
			responseTime.Seconds(), size, path)
		alive <- true
		return nil
	}
	// This doesn't seem to do what one exoects: FIXME?
	// reqerr, ok := err.(awserr.RequestFailure)
	//if ok {
	//	log.Printf("%s %f 0 0 %d %s %d GET\n",
	//		initial.Format("2006-01-02 15:04:05.000"),
	//		responseTime.Seconds(), size, path, reqerr.StatusCode)
	//	alive <- true
	// return nil
	//}
	fmt.Printf("%s %f 0 0 %d %s 4XX GET\n",
		initial.Format("2006-01-02 15:04:05.000"),
		responseTime.Seconds(), size, path)
	alive <- true
	return fmt.Errorf("unable to upload %q to %q, %v", path, bucket, err)
}

// mustCreateService creates a connection to an s3-compatible server.
func mustCreateService(myEndpoint string, awsLogLevel aws.LogLevelType) *s3.S3 {

	if S3params.AccessKey == "" {
		log.Fatal("called mustCreateService with no s3 params, internal error\n")
	}
	if verbose {
		awsLogLevel = aws.LogDebugWithSigning | aws.LogDebugWithHTTPBody
	}
	token := ""
	creds := credentials.NewStaticCredentials(S3params.AccessKey, S3params.SecretKey, token)
	_, err := creds.Get()
	if err != nil {
		log.Fatalf("bad credentials: %s\n", err)
	}
	cfg := aws.NewConfig().
		WithLogLevel(awsLogLevel).
		WithRegion("canada").
		WithEndpoint(myEndpoint).
		WithDisableSSL(true).
		WithS3ForcePathStyle(true).
		WithCredentials(creds)
	sess, err := session.NewSession() // There is a session.Must() for convenience
	if err != nil {
		log.Fatalf("bad session=%v\n", err)
	}
	svc = s3.New(sess, cfg)
	return svc
}

// AmazonS3Prep makes sure we have an amazon s3 session and any other prerequisites.
func AmazonS3Prep(baseURL string) {
	if svc == nil {
		svc = mustCreateService(baseURL, awsLogLevel)
	}
}
