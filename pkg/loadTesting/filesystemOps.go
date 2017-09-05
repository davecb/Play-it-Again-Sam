package loadTesting

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"
)

// TimedCreateFilesystemFile is for local (non-Protocol) file creation
func TimedCreateFilesystemFile(fullPath string, size int64) error {
	initial := time.Now() //               Response time starts
	MustCreateFilesystemFile(fullPath, size)
	responseTime := time.Since(initial) // Response time ends
	fmt.Printf("%s %f 0 0 %d %s 201 PUT\n",
		initial.Format("2006-01-02 15:04:05.000"),
		responseTime.Seconds(), size, fullPath)
	alive <- true
	return nil

}

// MustCreateFilesystemFile implements making the file in a filesystem relative to the current directory
// It's used by both local and s3.
func MustCreateFilesystemFile(fullPath string, size int64) {
	// fmt.Printf("in createFilesystemFile(%s, %s)\n", fullPath, size)
	dir := path.Dir(fullPath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatalf("could not create directories of %q, %v", fullPath, err)
	}
	out, err := os.Create(fullPath)
	if err != nil {
		log.Fatalf("could not create file %q, %v", fullPath, err)
	}
	in, err := os.Open("/dev/urandom")
	if err != nil {
		log.Fatalf("could not open /dev/urandom, %v", err)
	}
	defer in.Close() // nolint
	_, err = io.CopyN(out, in, size)
	if err != nil {
		log.Fatalf("could not close %q, %v", fullPath, err)
	}
	err = out.Close()
	if err != nil {
		log.Fatalf("error closing %q, %v", fullPath, err)
	}
}
