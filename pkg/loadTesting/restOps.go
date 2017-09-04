package loadTesting

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

// RestGet does a GET from an http target and times it
func RestGet(baseURL, path string) {
	req, err := http.NewRequest("GET", baseURL+"/"+path, nil)
	if err != nil {
		// try running right through this
		// log.Fatalf("error creating http request, %v: halting.\n", err)
		log.Printf("error creating http request, %v: continuing.\n", err)
		fmt.Printf("%s 0 0 0 0 %s %d GET\n",
			time.Now().Format("2006-01-02 15:04:05.000"), path, -1)
		alive <- true
		return
	}
	if verbose {
		var dump []byte
		dump, err = httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Fatalf("error dumping http request, %v: halting.\n", err)
		}
		fmt.Printf("Request: %s\n", dump)
	}

	initial := time.Now() // Response time starts
	resp, err := http.DefaultClient.Do(req)
	latency := time.Since(initial) // Latency ends
	if err != nil {
		//log.Fatalf("error getting http response, %v: halting.\n", err)
		// try running right through this
		log.Printf("error getting http response, %v: continuing.\n", err)
		fmt.Printf("%s %f %f 0 %d %s %d GET\n",
			initial.Format("2006-01-02 15:04:05.000"),
			latency.Seconds(), 0.0, 0, path, -2)
		alive <- true
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	transferTime := time.Since(initial) - latency // Transfer time ends
	defer resp.Body.Close()                       // nolint
	if err != nil {
		//log.Fatalf(`error reading http response, "%v": halting.\n`, err)
		// try running right through this
		log.Printf("error reading http response, %v: continuing.\n", err)
		fmt.Printf("%s %f %f 0 %d %s %d GET\n",
			time.Now().Format("2006-01-02 15:04:05.000"),
			latency.Seconds(), transferTime.Seconds(), 0, path, -3)
		alive <- true
		return
	}
	if verbose {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			// Error dumping http response, trying without body.
			// That avoids errors if length headers are wrong
			dump, err = httputil.DumpResponse(resp, false)
			if err != nil {
				log.Fatalf("error dumping http response, %v: halting.\n", err)
			}
		}
		fmt.Printf("Response: %s\n", dump)
	}

	fmt.Printf("%s %f %f 0 %d %s %d GET\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), transferTime.Seconds(), len(body), path, resp.StatusCode)
	alive <- true
}

// RestPut does an ordinary (not ceph or s3) put operation.
// FIXME add err back as a return value
func RestPut(baseURL, path string, size int64) {

	log.Printf("putting %s\n", baseURL+"/"+path)
	// make sure we have a dummy file
	if size <= 0 {
		fmt.Printf("%s 0 0 0 %d %s %d PUT\n",
			time.Now().Format("2006-01-02 15:04:05.000"),
			size, path, 411) // 411 means "length required"
		alive <- true
		return
	}
	fp, err := os.Open(junkDataFile)
	if err != nil {
		log.Fatalf("Can't open %q\n", junkDataFile)
	}
	defer fp.Close() // nolint

	initial := time.Now() // Response time starts
	// do put
	req, err := http.NewRequest("PUT", baseURL+"/"+path, io.LimitReader(fp, size))
	if err != nil {
		log.Fatalf("error creating http request, %v: halting.\n", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Timeouts and bad parameters will trigger this case. FIXME, continue?
		log.Fatalf("error getting http response, %v: halting.\n", err)
	}
	latency := time.Since(initial) // Response time ends
	contents, err := ioutil.ReadAll(resp.Body)
	transferTime := time.Since(initial) - latency // Transfer time ends
	defer resp.Body.Close()                       // nolint
	if err != nil {
		// FIXME, continue?
		log.Fatalf(`error reading http response, "%v": halting.\n`, err)
	}
	if verbose {
		//dump, err := httputil.DumpResponse(resp, true)
		//contents, err := ioutil.ReadAll(response.Body)
		//if err != nil {
		//	log.Fatalf("error dumping http response, %v: halting.\n", err)
		//}
		fmt.Print("Response:\n")
		fmt.Printf("    Length: %d\n", len(string(contents)))
		fmt.Printf("    Url: %q\n", baseURL+"/"+path)
		shortDescr, _ := codeDescr(resp.StatusCode)
		fmt.Printf("    Status code: %d %s\n", resp.StatusCode, shortDescr)
		hdr := resp.Header
		for key, value := range hdr {
			fmt.Println("   ", key, ":", value)
		}
		fmt.Printf("    Contents: %s\n", string(contents))
	}

	fmt.Printf("%s %f %f 0 %d %s %d PUT\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), transferTime.Seconds(), len(contents), path, resp.StatusCode)
	alive <- true
}
