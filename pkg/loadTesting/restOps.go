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

// RestProto satisfies operation by doing rest operations.
type RestProto struct {
	prefix string
}

// Init does nothing
func (p RestProto) Init() {}

// Tuning for large loads. With this, when we start reporting network
// errors, then we've overloaded somebody. Previously we bottlenecked
// on closed but not recycled sockets/
// For calvin this doesn't change the results up to 240, but
// then the former setup failed.
const (
	MaxIdleConnections int = 100
	RequestTimeout     int = 0
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost: MaxIdleConnections,
	},
	Timeout: time.Duration(RequestTimeout) * time.Second,
}

// Get does a GET from an http target and times it
func (p RestProto) Get(path string) error {
	if conf.Debug {
		log.Printf("in rest.Get(%s)\n", path)
	}
	req, err := http.NewRequest("GET", p.prefix+"/"+path, nil)
	if err != nil {
		dumpXact(req, nil, nil, conf.Crash, "error creating http request", err)
		fmt.Printf("%s 0 0 0 0 %s %d GET\n",
			time.Now().Format("2006-01-02 15:04:05.000"), path, -1)
		alive <- true
		return nil
	}

	if !conf.Cache {
		req.Header.Add("cache-control", "no-cache")
	}
	if conf.HostHeader != "" {
		req.Host = conf.HostHeader
		// Go disfeature: host is special,
		// See also https://github.com/golang/go/issues/7682
		req.Header.Add("Host", conf.HostHeader)
	}

	initial := time.Now() // Response time starts
	resp, err := httpClient.Do(req)
	latency := time.Since(initial) // Latency ends
	if err != nil {
		//log.Fatalf("error getting http response, %v: halting.\n", err)
		// try running right through this
		dumpXact(req, resp, nil, conf.Crash, "error getting http response", err)
		fmt.Printf("%s %f %f 0 %d %s %d GET\n",
			initial.Format("2006-01-02 15:04:05.000"),
			latency.Seconds(), 0.0, 0, path, -2)
		alive <- true
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	transferTime := time.Since(initial) - latency // Transfer time ends
	defer resp.Body.Close()                       // nolint
	if err != nil {
		dumpXact(req, resp, body, conf.Crash,"error reading http response, continuing", err)
		fmt.Printf("%s %f %f 0 %d %s %d GET\n",
			time.Now().Format("2006-01-02 15:04:05.000"),
			latency.Seconds(), transferTime.Seconds(), 0, path, -3)
		alive <- true
		return nil
	}
	// And, in the non-error cases, conditionally dump
	switch {
	case badGetCode(resp.StatusCode):
		dumpXact(req, resp, body, conf.Crash,"bad return code", nil)
	case badLen(resp.ContentLength, body):
		dumpXact(req, resp, body,  conf.Crash,"bad length", nil)
	case conf.Verbose:
		dumpXact(req, resp, body, conf.Crash,"", nil)
	}


	fmt.Printf("%s %f %f 0 %d %s %d GET\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), transferTime.Seconds(), len(body), path, resp.StatusCode)
	alive <- true
	return nil
}

// Put does an ordinary REST (not ceph or s3) put operation.
func (p RestProto) Put(path string, size int64) error {

	if conf.Debug {
		log.Printf("in rest.Pet(%s, %d)\n", path, size)
		//log.Printf("putting %s\n", p.prefix+"/"+path)
	}
	if size <= 0 {
		fmt.Printf("%s 0 0 0 %d %s %d PUT\n",
			time.Now().Format("2006-01-02 15:04:05.000"),
			size, path, 411) // 411 means "length required"
		alive <- true
		return nil
	}
	// make sure we have a dummy file
	fp, err := os.Open(junkDataFile)
	if err != nil {
		return fmt.Errorf("can't open %q", junkDataFile)
	}
	defer fp.Close() // nolint

	initial := time.Now() // Response time starts
	// do put
	req, err := http.NewRequest("PUT", p.prefix+"/"+path, io.LimitReader(fp, size))
	if err != nil {
		// report and exit
		dumpXact(req, nil, nil, true, "error creating http request", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		// Timeouts and bad parameters will trigger this case. FIXME, continue?
		dumpXact(req, nil, nil, true, "error getting http response", err)
	}
	latency := time.Since(initial) // Response time ends
	contents, err := ioutil.ReadAll(resp.Body)
	transferTime := time.Since(initial) - latency // Transfer time ends
	defer resp.Body.Close()                       // nolint
	if err != nil {
		dumpXact(req, resp, contents, true,"error reading http response", err)
	}
	// And, in the non-error cases, conditionally dump
	switch {
	case badGetCode(resp.StatusCode): // FIXME, putCode
		dumpXact(req, resp, contents, conf.Crash,"bad return code", nil)
	case conf.Verbose:
		dumpXact(req, resp, contents, conf.Crash,"", nil)
	}
	fmt.Printf("%s %f %f 0 %d %s %d PUT\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), transferTime.Seconds(), len(contents), path, resp.StatusCode)
	alive <- true
	return nil
}

// badGetCode is true if this isn't a 20X or 404
// in this case "bad" means "ucky"
func badGetCode(i int) bool {
	if i == 200 || i == 202 || i == 404 {
		return false
	}
	// if --crash is set, returning true will trigger
	// a dump of the bad transaction and a shutdown
	return true
}

// badLen is true if we have zero body lengths
func badLen(bodylen int64, body []byte) bool {
	if bodylen == 0 || len(body) == 0 {
		return true
	}
	return false
}

// dumpXact dumps request and response together, with a reason
func dumpXact(req *http.Request, resp *http.Response, body []byte, crash bool, reason string, err error) {
	var r string
	if err != nil {
		r = fmt.Sprintf("%s, %v\n", reason, err)
	} else {
		r = fmt.Sprintf("%s\n", reason)
	}
	r += requestToString(req)
	r += responseToString(resp)
	r += bodyToString(body)
	log.Printf("%s\n", r)
	if crash {
		log.Fatalf("halting.\n")
	}
}

// requestToString provides extra information about an http request if it can
func requestToString(req *http.Request) string {
	var dump []byte
	var err error

	if req == nil {
		return "Request: <nil>\n"
	}
	dump, err = httputil.DumpRequestOut(req, true)
	if err != nil {
		return fmt.Sprintf("fatal error dumping http request, %v\n", err)
	}
	return fmt.Sprintf("Request: \n%s", dump)
}

// responseToString provides extra information about an http response
func responseToString(resp *http.Response) string {
	if resp == nil {
		return"Response: <nil>\n"
	}
	contents, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return fmt.Sprintf("error dumping http response, %v\n", err)
	}
	s := "Response headers:\n"
	s += fmt.Sprintf("    Length: %d\n", resp.ContentLength)
	shortDescr, _ := codeDescr(resp.StatusCode)
	s += fmt.Sprintf("    Status code: %d %s\n", resp.StatusCode, shortDescr)
	hdr := resp.Header
	for key, value := range hdr {
		s += fmt.Sprintln("   ", key, ":", value)
	}
	s += fmt.Sprintf("Response contents: \n%s", string(contents))
	return s
}

// bodyToString
func bodyToString(body []byte) string {
	if body == nil {
		return "Body: <nil>\n"
	}
	return fmt.Sprintf("Body:\n %s\n", body)
}
