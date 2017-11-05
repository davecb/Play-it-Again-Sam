# runLoadTest(1) 
runLoadTest - replay a load against a new target
## SYNOPSIS
 Usage: runLoadTest --tps TPS [--progress TPS][...][-v] load-file.csv baseURL
file URL

## DESCRIPTION
This program runs a load test from a journal file, and records its results in 
a file or journal with the same format. It is typically used in load testing to 
to replay a set of operations at a higher load.

For example, if a typical hour's load was collected from a web 
server logs, it could be replayed at a different load by saying
`runLoadTest -tps 10 file url`  or at increasing loads
with `runLoadTest -tps 50 -progression 10 file url`.
The latter would run it at 10 TPS, then 20, 30 and so on, up to 100 TPS.

There are a large number of options, each of which is described below. 
Either a single or double-dash can be used with any option. 
Eg, --tps and -tps mean the same thing.

### Load options    
--tps int  
*  TPS target (required)    
   This option set the maximum load in transactions 
   per second (requests and responses per second) 
   If it is used alone, the test will run at that
   TPS until done. This is for the classic long-term
   run test.
      
-progress int 
* progress rate, in TPS steps   
  If this option is used, the load will be progressively
  increased by this amount, until it reaches the tps 
  target. As noted above `-tps 50 -progression 10` will 
  start with a load of 10 TPS, then 20, 30, 40 and 50.
  This is used to find the performance at increasing load
  and find the inflection point in the "_/" hockey-stick
  curve.
  
-duration int 
* Duration of a step (default 10)   
  When doing progressions, this is the length of a step in seconds.
  Some applications fail after a certain number of requests, or one
  may have a limited-size file, so this allows one to shorten
  (or lengthen) the tests at any given speed.
  
-start-tps int   
* TPS to start from   
  If specified, this will be the initial load in TPS. 
  For example, `-tps 50 -progression 10 --start-tps 20`
  will start with a load of 20 TPS, then 30, 40 and 50.
  This is handy when one has already done a test at a low range of TPS
  and wishes to test at higher loads.
     
-real-time 
* Override -tps and tail -f the input file instead.    
  This allows a machine to be fed the same load as another machine
  at the same time. It is for parallel running and finding cases
  where the new program differs from the old.
      

### Data options   
-for int 
* number of records to use, eg 1000.   
  This limits the length of the run to a specific number of records
  from the input file.  

-from int 
* number of records to skip, eg 100.   
  This starts at a particular record.
  
  These are for doing limited tests, or for doing tests that behave 
  differently between the first and subsequent repetitions, such
  as test of caches.   

### Protocol options    
-rest 
* use rest protocol 
  Do GETs as unauthenticated REST calls.
   
-s3 
* use s3 protocol
  Do GETS as authenticated s3 calls
  
  The default is to do GETs only: PUTs and DELEs are currently disabled,
  but have been used experimentally and will be refactored and enabled
  later.  POSTs are deferred until I get a good example to develop 
  a use case from. 

### S3 options     
-s3-bucket string 
* set bucket when using s3 protocol  
  This allows multiple "buckets" of files from a single logical host.
    
-s3-key string 
* set key when using s3 protocol
  This is the equivalent of an application-id or user-name in s3
  (default "KEY NOT SET")   
        
-s3-secret string 
* set secret when using s3 protocol 
  This is the equivalent to a password (default "SECRET NOT SET")     

  These are typically set in a configuration file (see below) as
  they do not change often. Command-line options override the
  configuration file.

###Convenience options          
-strip string 
* text to strip from paths 
  This is for removing prefixes that appear in the input. If stripped,
  they will not appear in the output file. 
   
-cache 
* allow caching  
  Normally a no-cache header is sent: this disables it. 
      
-host-header string 
* add a Host: header 
  Some sites require a host header (eg, when you are using an IP address
  in the URL). This sets it  
  
-serialize 
* serialize load 
  This is for limiting the number of requests outstanding, by skipping
   requests while waiting for responses. This creates an invalid test, 
   but is useful for providing a limited load during debugging.   
   
-save 
* save downloaded file(s) as ./loadTest.out
  This attempts to save the response from a GET as a file 
  in the current directory. Not thread-safe, so only use for
  debugging with -tps 1 -for 1
   
-d	
* add debugging messages  
  This is for debugging the load generator itself.
      
-v
* add verbose messages    
  This is for debugging the system under test, by seeing more about
  what it is doing. Shows the request and response in more detail.

### Config-file options 
These options are from the config-file parser, which allows any of the
above options to be specified in a configuration file.
   
-allowMissingConfig 
 * Don't terminate the app if the ini file cannot be read. 
   
-allowUnknownFlags 
 * Don't terminate the app if ini file contains unknown flags.  
 
-config string 
 * Path to ini config for using in go flags. May be relative to the 
 current executable path.   
 
-configUpdateInterval duration 
* Update interval for re-reading config file set via -config flag. 
  Zero disables config file re-reading. 
   
-dumpflags 
* Dumps values for all flags defined in the app into stdout in 
  ini-compatible syntax and terminates the app.    


## FILES
The input and output files are identical, of the form
```csv
#yyy-mm-dd hh:mm:ss latency xfertime thinktime bytes url rc op
2017-09-21 08:15:07.270 0 0 0 0 /upload/images/383bcc59-354b-46fb-b66c-0907b21fad94_albert.jpg 200 GET

```
As an input, only the url is significant. It is concatenated with the 
url prefix provide on the command-line and sent.

As an output, the analyzable fields are
* latency   
  This is the time between the request and the first byte(s) of the 
  response. It is the "wait" before the response shows up. It includes
  the network time, the time it took to process the request, and the
  wait in queue if the server could not process the request immediately.
  
* xfertime  
  This is the time between the first byte of the response and the end.
  It is the network time it takes to transfer the data returned. It may
  be zero if all the data arrives in the first packet.
   
* thinktime   
  This is the time between the end of a request and the beginning of the 
  next one, which is an indication of a human's think time when measuring
  user-provided loads. It is zero in load-testing use.
  
* bytes     
  This is the number of bytes sent during the transfer time.
  
* rc    
  This is the http return code
  
* op   
  This is the REST operation, currently limited to GETs
 

## "SEE ALSO"
perf2seconds(1), nginx2perf(1), mkLoadTestFiles(1)

## EXAMPLES
This is a test of my storage machine, _calvin_, from
10 to 250 TPS in steps of 10 TPS. The output is
sent to `perf2seconds` to report on 10-second 
samples

```bash

runLoadTest --rest  --tps 250 --progress 10 \
    --duration 30 \
	../sample.csv http://calvin > raw.csv
perf2seconds raw.csv >calvin10to250.csv
 
```

## BUGS
PUT and DELE require refactoring and have been disabled.

^C kills everything instantly. 

Instead of a put test for filesystems, a separate program called
`mkLoadTestFiles` creates files of the required sizes.

## DIAGNOSTICS
If an error occurs, if an unexpected return code is 
received (eg, a 503) or if -v is specified, the request and response
will be written to stderr.

During normal operation, a small number of status messages will also
be written to stderr to indicate the progress of the test.  


## AUTHOR

David Collier-Brown
