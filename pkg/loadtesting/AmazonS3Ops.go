package loadtesting

// AmazonS3Ops implements s3 get, put and delete using the Amazon client library.
// Initially the Amazon library was too buggy, but Marcus Watt of the ceph
// team debugged it for me. I expect most people will use the Amazon library.

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Proto satisfies operation by doing rest operations.
type S3Proto struct {
	prefix string
}

var svc *s3.S3
var awsLogLevel = aws.LogOff

// Get does a get operation from an s3Protocol target and times it,
func (p S3Proto) Get(path string, oldRc string) {
	if conf.Debug {
		log.Printf("in AmazonS3Get(%s, %s)\n", p.prefix, path)

		head, err := svc.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(conf.S3Bucket),
			Key:    aws.String(path),
		})
		if err != nil {
			fmt.Println("HeadObject err", err)
		} else {
			fmt.Println("HeadObject ", head)
			//HeadObject  {
			//	AcceptRanges: "bytes",
			//	ContentLength: 7623,
			//	ContentType: "text/plain",
			//	ETag: "\"8d91918d01db7e2e959e6dbf4d0cd646\"",
			//	LastModified: 2017-10-30 13:19:21 +0000 UTC,
			//	Metadata: {
			//	Ipaddresses: "10.92.10.201, 10.92.10.201"
			//}
		}
	}

	file, err := os.CreateTemp("/tmp", "loadTesting")
	if err != nil {
		log.Fatalf("Unable to create a temp file,  %v", err)
	}
	defer os.Remove(file.Name()) // nolint

	downloader := s3manager.NewDownloaderWithClient(svc)
	initial := time.Now() //              				***** Response time starts
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(conf.S3Bucket),
			Key:    aws.String(path),
		})
	responseTime := time.Since(initial) // 				***** Response time ends
	if err != nil {
		rc := errorCodeToHTTPCode(err)
		fmt.Printf("%s %f 0 0 %d %s %d GET\n",
			initial.Format("2006-01-02 15:04:05.000"),
			responseTime.Seconds(), numBytes, path, rc)
		reportPerformance(initial, responseTime, 0, nil, path, rc, oldRc)
		return
	}
	fmt.Printf("%s %f 0 0 %d %s 200 GET\n",
		initial.Format("2006-01-02 15:04:05.000"),
		responseTime.Seconds(), numBytes, path)
	reportPerformance(initial, responseTime, 0, nil, path, 200, oldRc)
}

// Put puts a file and times it
// error return is used only by mkLoadTestFiles  FIXME
func (p S3Proto) Put(path, size, oldRC string) {
	log.Fatalf("put is not implemented yet\n")
	//if conf.Debug {
	//	log.Printf("in AmazonS3Put(%s, %s, %d)\n", p.prefix, path, size)
	//}
	//
	//file, err := os.Open(junkDataFile)
	//if err != nil {
	//	return fmt.Errorf("Unable to open junk-data file %s, %v", junkDataFile, err)
	//}
	//defer file.Close() // nolint
	//lr := io.LimitReader(file, size)
	//
	//if svc == nil {
	//	return fmt.Errorf("missing service %v", svc)
	//}
	//uploader := s3manager.NewUploaderWithClient(svc)
	//initial := time.Now() //              				***** Response time starts
	//_, err = uploader.Upload(&s3manager.UploadInput{
	//	Bucket: aws.String(conf.S3Bucket),
	//	Key:    aws.String(path),
	//	Body:   lr,
	//})
	//responseTime := time.Since(initial) // 				***** Response time ends
	//// FIXME swap this around
	//if err == nil {
	//	fmt.Printf("%s %f 0 0 %d %s 201 PUT\n",
	//		initial.Format("2006-01-02 15:04:05.000"),
	//		responseTime.Seconds(), size, path)
	//	return nil
	//}
	//// This doesn't seem to do what one expects: FIXME?
	//// reqerr, ok := err.(awserr.RequestFailure)
	////if ok {
	////	log.Printf("%s %f 0 0 %d %s %d GET\n",
	////		initial.Format("2006-01-02 15:04:05.000"),
	////		responseTime.Seconds(), size, path, reqerr.StatusCode)
	//// return nil
	////}
	//fmt.Printf("%s %f 0 0 %d %s 4XX GET\n",
	//	initial.Format("2006-01-02 15:04:05.000"),
	//	responseTime.Seconds(), size, path)
	//return fmt.Errorf("unable to upload %q to %q, %v", path, conf.S3Bucket, err)
}

// Post for s3: not implemented yes
func (p S3Proto) Post(path, size, oldRC, body string) {
	log.Fatalf("POST is unimplemented\n")
}

// mustCreateService creates a connection to an s3-compatible server.
func mustCreateService(myEndpoint string, awsLogLevel aws.LogLevelType) *s3.S3 {

	if conf.S3Key == "" {
		log.Fatal("called mustCreateService with no s3 params, internal error\n")
	}
	if conf.Verbose {
		awsLogLevel = aws.LogDebugWithSigning | aws.LogDebugWithHTTPBody |
			aws.LogDebugWithRequestErrors
	}
	token := ""
	creds := credentials.NewStaticCredentials(conf.S3Key, conf.S3Secret, token)
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

// Init makes sure we have an amazon s3 session and any other prerequisites.
func (p S3Proto) Init() {
	if svc == nil {
		svc = mustCreateService(p.prefix, awsLogLevel)
	}
}

// errorCodeToHTTPCode is wimpey!
// only a few codes (eg, 404) are implemented
func errorCodeToHTTPCode(err error) int {
	aerr, ok := err.(awserr.Error)
	if !ok {
		return -2 // not from aws
	}
	reqErr, ok := aerr.(awserr.RequestFailure)
	if !ok {
		return -1 // not a request failure
	}
	// A service error occurred, it has an HTTP code
	return reqErr.StatusCode()
}
