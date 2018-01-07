package loadTesting

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
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
func (p RestProto) Get(path string, oldRc string) error {
	if conf.Debug {
		log.Printf("in rest.Get(%s)\n", path)
	}
	req, err := http.NewRequest("GET", p.prefix+"/"+path, nil)
	if err != nil {
		dumpXact(req, nil, nil, conf.Crash, "error creating http request", err)
		reportPerformance(time.Now(), 0, 0, nil, path, -1, oldRc)
		alive <- true
		return nil
	}
	addHeaders(req)

	initial := time.Now() // Response time starts
	resp, err := httpClient.Do(req)
	latency := time.Since(initial) // Latency ends
	if err != nil {
		//log.Fatalf("error getting http response, %v: halting.\n", err)
		// try running right through this
		dumpXact(req, resp, nil, conf.Crash, "error getting http response", err)
		// 444 is nginx's code for server has returned no information ans/or EOF
		reportPerformance(initial, latency, 0, nil, path, 444, oldRc)
		alive <- true
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	// how about io.Copy(ioutil.Discard, resp.Body)
	transferTime := time.Since(initial) - latency // Transfer time ends
	defer resp.Body.Close()                       // nolint
	if err != nil {
		dumpXact(req, resp, body, conf.Crash, "error reading http response, continuing", err)
		// the resp is available, the body, distinctly less so (;-))
		reportPerformance(initial, latency, transferTime, body, path, resp.StatusCode, oldRc)
		alive <- true
		return nil
	}

	// And, in the non-error cases, conditionally dump
	switch {
	case badGetCode(resp.StatusCode):
		dumpXact(req, resp, body, conf.Crash, "bad return code", nil)
	//case badLen(resp.ContentLength, body):
	//	dumpXact(req, resp, body, conf.Crash, "bad length", nil)
	case conf.Verbose:
		dumpXact(req, resp, body, conf.Crash, "verbose", nil)
	}

	reportPerformance(initial, latency, transferTime, body, path, resp.StatusCode, oldRc)
	alive <- true
	return nil
}

func addHeaders(req *http.Request) {
	if !conf.Cache {
		req.Header.Add("cache-control", "no-cache")
	}
	if conf.HostHeader != "" {
		req.Host = conf.HostHeader
		// Go disfeature: host is special,
		// See https://github.com/golang/go/issues/7682
		req.Header.Add("Host", conf.HostHeader)
	}
	if conf.AkamaiDebug {
		req.Header.Add("Pragma",
			"akamai-x-cache-on, "+
				"akamai-x-cache-remote-on, "+
				"akamai-x-check-cacheable, "+
				"akamai-x-get-cache-key, "+
				"akamai-x-get-ssl-client-session-id, "+
				"akamai-x-get-true-cache-key, "+
				"akamai-x-get-request-id")
	}
	for key, value := range conf.HeaderMap {
		req.Header.Add(key, value)
	}
}

// Put does an ordinary REST (not ceph or s3) put operation.
func (p RestProto) Put(path, size, oldRC string) error {
	var bytes int64
	var err error

	if conf.Debug {
		log.Printf("in rest.Put(%s, %s)\n", path, size)
		//log.Printf("putting %s\n", p.prefix+"/"+path)
	}
	bytes, err = strconv.ParseInt(size, 10, 64)
	if err != nil {
		bytes = 0 // FIXME
	}
	if bytes <= 0 {
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
	req, err := http.NewRequest("PUT", p.prefix+"/"+path, io.LimitReader(fp, bytes))
	if err != nil {
		// reportPerformance and exit
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
		dumpXact(req, resp, contents, true, "error reading http response", err)
	}
	// And, in the non-error cases, conditionally dump
	switch {
	case badGetCode(resp.StatusCode): // FIXME, putCode
		dumpXact(req, resp, contents, conf.Crash, "bad return code", nil)
	case conf.Verbose:
		dumpXact(req, resp, contents, conf.Crash, "", nil)
	}
	//reportPerformance(initial, latency, transferTime, body, path, resp, oldRc)
	fmt.Printf("%s %f %f 0 %s %s %d PUT\n",
		time.Now().Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), transferTime.Seconds(), size, path, resp.StatusCode)
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
//func badLen(bodylen int64, body []byte) bool {
//	if bodylen == 0 || len(body) == 0 {
//		return true
//	}
//	return false
//}

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
		return "Response: <nil>\n"
	}
	contents, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return fmt.Sprintf("error dumping http response, %v\n", err)
	}
	s := "Response information:\n"
	s += fmt.Sprintf("    Length: %d\n", resp.ContentLength)
	s += fmt.Sprintf("    Status code: %d %s\n", resp.StatusCode,
		http.StatusText(resp.StatusCode))
	//hdr := resp.Header
	//for key, value := range hdr {
	//	s += fmt.Sprintln("   ", key, ":", value) // FIXME remove []
	//}
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
